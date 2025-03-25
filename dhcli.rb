# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Dhcli < Formula
  desc "A command-line tool for DigitalHub platform."
  homepage "https://scc-digitalhub.github.io/"
  version "0.10.0-beta-gr"

  on_macos do
    url "https://github.com/scc-digitalhub/digitalhub-cli/releases/download/0.10.0-beta-gr/dhcli-darwin.tar.gz"
    sha256 "7d25193bf931283e055410e4514b1db5813b3c0aeff769493d80666d60c7ca0c"

    def install
      bin.install "dhcli"
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      if Hardware::CPU.is_64_bit?
        url "https://github.com/scc-digitalhub/digitalhub-cli/releases/download/0.10.0-beta-gr/dhcli-linux-amd64.tar.gz"
        sha256 "ae580511d1ea094f45ee235997896f8cffb0c140dfdee6b11712e790c1d55606"

        def install
          bin.install "dhcli"
        end
      end
    end
    if Hardware::CPU.arm?
      if Hardware::CPU.is_64_bit?
        url "https://github.com/scc-digitalhub/digitalhub-cli/releases/download/0.10.0-beta-gr/dhcli-linux-arm64.tar.gz"
        sha256 "dab86ef51c52e21f053feb17fcb8de2d86918ab4ccc1dbe86299f6494145ad5b"

        def install
          bin.install "dhcli"
        end
      end
    end
  end
end
