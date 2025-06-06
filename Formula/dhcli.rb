# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Dhcli < Formula
  desc "A command-line tool for DigitalHub platform."
  homepage "https://scc-digitalhub.github.io/"
  version "0.11.0"

  on_macos do
    url "https://github.com/scc-digitalhub/digitalhub-cli/releases/download/0.11.0/dhcli-darwin.tar.gz"
    sha256 "cecb5a267608e6d839f91087bc256902cbf2e75fdbfaf957a44cd42847bffffa"

    def install
      bin.install "dhcli"
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      if Hardware::CPU.is_64_bit?
        url "https://github.com/scc-digitalhub/digitalhub-cli/releases/download/0.11.0/dhcli-linux-amd64.tar.gz"
        sha256 "78e67fbe3cc8cbb77a635df36c35845ece543faa58a2ecf2eb9152d8296523a6"

        def install
          bin.install "dhcli"
        end
      end
    end
    if Hardware::CPU.arm?
      if Hardware::CPU.is_64_bit?
        url "https://github.com/scc-digitalhub/digitalhub-cli/releases/download/0.11.0/dhcli-linux-arm64.tar.gz"
        sha256 "eb4cbe6f8c8a9a99702c54cccc5575d89062451275ddeb7192c8717c0ae5627f"

        def install
          bin.install "dhcli"
        end
      end
    end
  end
end
