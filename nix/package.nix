{
  lib,
  stdenv,
  fetchFromGitHub,
  buildGoModule,
  ffmpeg,
  makeWrapper,
}: let
  version = "0.3.4";
in
  buildGoModule {
    pname = "radiogogo";
    inherit version;

    src = fetchFromGitHub {
      owner = "matteo-pacini";
      repo = "RadioGoGo";
      rev = "v${version}";
      hash = "sha256-StbSd6tHoe4hBQV2CMSBgHkgGhAINBkd6KyP1SlGi6k=";
    };

    vendorHash = "sha256-yOc1qYy3Qy4rpaSycPqWheOFW+mIrmzEuw3l+yZ0TRY=";

    nativeBuildInputs = [makeWrapper];

    ldflags = [
      "-s"
      "-w"
    ];

    postInstall = ''
      wrapProgram $out/bin/radiogogo \
          --prefix PATH : ${lib.makeBinPath [ffmpeg]}
    '';

    meta = with lib; {
      homepage = "https://github.com/matteo=pacini/RadioGoGo";
      description = "Go-powered CLI to surf global radio waves via a sleek TUI";
      license = licenses.mit;
      maintainers = with maintainers; [matteopacini];
      mainProgram = "radiogogo";
    };
  }
