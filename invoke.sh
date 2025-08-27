#!/usr/bin/env sh
# POSIX sh - invoke con upload diretto di ./output via dhcli

set -eu

# usage
if [ -z "${1-}" ]; then
  echo "Uso: INVOKE.sh <comando> [args...]" >&2
  exit 64
fi

COMMAND=$1
shift  # "$@" sono gli argomenti del comando target

# debug
if [ "${INVOKE_DEBUG-}" = "1" ]; then
  {
    printf "[INVOKE] Comando   : %s\n" "$COMMAND"
    printf "[INVOKE] Argomenti :"
    for a in "$@"; do
      printf " [%s]" "$a"
    done
    printf "\n"
  } >&2
fi

# logging (opzionale)
if [ -n "${INVOKE_LOG-}" ]; then
  TS="$(date -u +'%Y-%m-%dT%H:%M:%SZ')"
  line="$TS | CMD: $COMMAND"
  for a in "$@"; do line="$line $a"; done
  { echo "$line" >> "$INVOKE_LOG"; } || true
fi

# dry-run (opzionale)
if [ "${INVOKE_DRYRUN-}" = "1" ]; then
  exit 0
fi

# esecuzione comando target
"$COMMAND" "$@"
EXIT_CODE=$?

# ---- BLOCCO UPLOAD ./output → artifact ----
# Condizioni:
#   - Se ./output esiste ed è NON vuota, si tenta l'upload con:
#     ./dhcli upload artifact -n <ARTIFACT_NAME> -f ./output -p <PROJECT_NAME>
#   - PROJECT_NAME deve essere presente in env per procedere con l'upload.
#   - ARTIFACT_NAME = "$RUN_ID-OUTPUT" (oppure timestamp se RUN_ID mancante).
# Exit policy:
#   - Se EXIT_CODE=0 e l'upload fallisce -> esci con l'exit code della CLI (EXIT_CODE_CLI)
#   - Altrimenti -> esci con EXIT_CODE

EXIT_CODE_CLI=0

if [ -d "./output" ]; then
  # verifica che non sia vuota
  if find "./output" -mindepth 1 -print -quit 2>/dev/null | grep -q .; then
    # serve PROJECT_NAME dall'ambiente
    if [ -z "${PROJECT_NAME-}" ]; then
      echo "[INVOKE] Variabile d'ambiente PROJECT_NAME non impostata: impossibile fare upload." >&2
      EXIT_CODE_CLI=2
    else
      # nome artefatto
      if [ -n "${RUN_ID}" ]; then
        ARTIFACT_NAME="${RUN_ID}-OUTPUT"
      else
        ARTIFACT_NAME="$(date -u +%Y%m%dT%H%M%SZ)-OUTPUT"
      fi

      # debug
      if [ "${INVOKE_DEBUG-}" = "1" ]; then
        printf "[INVOKE] Upload artifact: name=%s project=%s path=./output\n" "$ARTIFACT_NAME" "$PROJECT_NAME" >&2
      fi

      # esegui upload diretto della directory ./output
      ./dhcli upload artifact -n "$ARTIFACT_NAME" -f ./output -p "$PROJECT_NAME"
      EXIT_CODE_CLI=$?
    fi
  fi
fi

# logging exit del comando target
if [ -n "${INVOKE_LOG-}" ]; then
  { printf '%s | EXIT: %d\n' "$(date -u +'%Y-%m-%dT%H:%M:%SZ')" "$EXIT_CODE" >> "$INVOKE_LOG"; } || true
fi

# regola di uscita finale
if [ "$EXIT_CODE" -eq 0 ] && [ "${EXIT_CODE_CLI:-0}" -ne 0 ]; then
  exit "$EXIT_CODE_CLI"
fi
exit "$EXIT_CODE"
