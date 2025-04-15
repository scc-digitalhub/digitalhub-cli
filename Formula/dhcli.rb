# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Dhcli < Formula
  desc "A command-line tool for DigitalHub platform."
  homepage "https://scc-digitalhub.github.io/"
  version "0.10.3"

  on_macos do
    url "https://github.com/scc-digitalhub/digitalhub-cli/releases/download/0.10.3/dhcli-darwin.tar.gz"
    sha256 "71b77dc2a6204c1742fce797c3e8fb95ca156134dd7d9fa059cf289da69b6f61"

    def install
      bin.install "dhcli"
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      if Hardware::CPU.is_64_bit?
        url "https://github.com/scc-digitalhub/digitalhub-cli/releases/download/0.10.3/dhcli-linux-amd64.tar.gz"
        sha256 "8b0ac62ab01779dfe27f8ae45a018fafc3b9941c20e2f79167228c073fa9ebca"

        def install
          bin.install "dhcli"
        end
      end
    end
    if Hardware::CPU.arm?
      if Hardware::CPU.is_64_bit?
        url "https://github.com/scc-digitalhub/digitalhub-cli/releases/download/0.10.3/dhcli-linux-arm64.tar.gz"
        sha256 "b95153c2e4114d2e68a9c99d2253ea8c9ff7b5bfd834c7d4a13640b60755e63d"

        def install
          bin.install "dhcli"
        end
      end
    end
  end
end
