#!/usr/bin/env bash
set -Eeuo pipefail

# ----------------------------------
# Usage:
#   ./ultra-upload.sh <project>
#   PROJECT=my-project N=500 J=4 ./ultra-upload.sh
#
# Env override:
#   N                = numero di upload (default 1000)
#   J                = job in parallelo (default 4)
#   SLEEP_MS         = pausa minima tra richieste (ms, default 150)
#   SLEEP_JITTER_MS  = jitter extra 0..JITTER (ms, default 150)
#   RETRIES          = tentativi/oggetto (default 3)
#   BASE_BACKOFF_MS  = backoff iniziale (ms, default 300)
#   MAX_BACKOFF_MS   = backoff massimo (ms, default 5000)
#   KEEP_FILES       = 1 per NON cancellare i file (default 0)
#   ARTIFACT_PREFIX  = prefisso per -n (default "test")
# ----------------------------------

# ---- Parametri & config ----
PROJECT="${PROJECT:-${1:-}}"
if [[ -z "$PROJECT" ]]; then
  echo "Uso: $0 <project>  (oppure PROJECT=... $0)" >&2
  exit 64
fi

: "${N:=1000}"
: "${J:=4}"
: "${SLEEP_MS:=150}"
: "${SLEEP_JITTER_MS:=150}"
: "${RETRIES:=3}"
: "${BASE_BACKOFF_MS:=300}"
: "${MAX_BACKOFF_MS:=5000}"
: "${KEEP_FILES:=0}"
: "${ARTIFACT_PREFIX:=test}"

TMPDIR="$(mktemp -d -t dhcli-upload.XXXXXXXX)"
LOCKFILE="$TMPDIR/upload.lock"

cleanup() {
  if [[ "$KEEP_FILES" = "1" ]]; then
    echo "File generati conservati in: $TMPDIR"
  else
    rm -rf "$TMPDIR" || true
  fi
}
trap cleanup EXIT INT TERM

# dipendenze opzionali
have_flock=0
command -v flock >/dev/null 2>&1 && have_flock=1

# esporta per le subshell xargs
export TMPDIR LOCKFILE PROJECT N J SLEEP_MS SLEEP_JITTER_MS RETRIES BASE_BACKOFF_MS MAX_BACKOFF_MS KEEP_FILES ARTIFACT_PREFIX have_flock

# generatori
gen_hex()   { od -vN8 -An -tx1 /dev/urandom | tr -d ' \n'; }
gen_uuid()  { command -v uuidgen >/dev/null 2>&1 && uuidgen || echo ""; }
export -f gen_hex gen_uuid

upload_one() {
  set -Eeuo pipefail
  local i="$1"

  # --- crea file univoco con contenuto random ---
  local ts rhex f
  ts="$(date +%s%N)"
  rhex="$(gen_hex)"
  f="$(mktemp -p "$TMPDIR" "prova_${ts}_${rhex}_XXXXXXXX.txt")"

  {
    printf 'index=%s\n' "$i"
    printf 'created_at=%s\n' "$(date -Iseconds)"
    printf 'rand_hex=%s\n' "$rhex"
    head -c 256 /dev/urandom | base64 | tr -d '\n'
    printf '\n'
  } >"$f"

  # --- nome artifact randomico ---
  local u a_name
  u="$(gen_uuid)"
  if [[ -n "$u" ]]; then
    a_name="${ARTIFACT_PREFIX}-${u}"
  else
    a_name="${ARTIFACT_PREFIX}-${ts}-${rhex}"
  fi

  # pacing calcolato qui (base + jitter)
  local jitter sleep_ms sleep_sec
  jitter=$(( RANDOM % (SLEEP_JITTER_MS + 1) ))
  sleep_ms=$(( SLEEP_MS + jitter ))
  sleep_sec="$(printf '%d.%03d' $((sleep_ms/1000)) $((sleep_ms%1000)))"

  # retry + backoff con lock globale per spaziatura fra richieste
  local rc attempt=1
  while :; do
    # ---- sezione critica: 1 solo upload alla volta + mini sleep ----
    if [[ "$have_flock" -eq 1 ]]; then
      exec 200>>"$LOCKFILE"
      flock -x 200
    else
      while ! mkdir "$LOCKFILE.d" 2>/dev/null; do sleep 0.05; done
    fi

    set +e
    ./dhcli upload artifact -n "$a_name" -f "$f" -p "$PROJECT"
    rc=$?
    set -e

    # pausa globale (anche se fallisce)
    sleep "$sleep_sec"

    if [[ "$have_flock" -eq 1 ]]; then
      flock -u 200
      exec 200>&-
    else
      rmdir "$LOCKFILE.d" 2>/dev/null || true
    fi
    # ---- fine sezione critica ----

    if [[ $rc -eq 0 ]]; then
      printf 'OK   %s -> %s (artifact=%s project=%s)\n' "$i" "$f" "$a_name" "$PROJECT"
      break
    fi
    if [[ $attempt -ge $RETRIES ]]; then
      printf 'FAIL %s -> %s (artifact=%s project=%s) after %d attempts\n' "$i" "$f" "$a_name" "$PROJECT" "$attempt"
      break
    fi

    # backoff esponenziale + jitter
    local back_ms=$(( BASE_BACKOFF_MS * (1 << (attempt - 1)) ))
    (( back_ms > MAX_BACKOFF_MS )) && back_ms="$MAX_BACKOFF_MS"
    local j_ms=$(( RANDOM % (BASE_BACKOFF_MS + 1) ))
    local total_ms=$(( back_ms + j_ms ))
    local back_sec
    back_sec="$(printf '%d.%03d' $((total_ms/1000)) $((total_ms%1000)))"
    sleep "$back_sec"

    attempt=$((attempt+1))
  done

  [[ "$KEEP_FILES" = "1" ]] || rm -f "$f" || true
}
export -f upload_one

# ---- Avvio worker ----
seq 1 "$N" | xargs -r -n1 -P "$J" -I{} bash -c 'upload_one "$@"' _ {}

