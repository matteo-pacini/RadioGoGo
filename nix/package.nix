{
  lib,
  stdenv,
  fetchFromGitHub,
  buildGoModule,
  ffmpeg,
  makeWrapper,
}: let
  version = "0.3.3";
in
  buildGoModule {
    pname = "radiogogo";
    inherit version;

    src = fetchFromGitHub {
      owner = "matteo-pacini";
      repo = "RadioGoGo";
      rev = "v${version}";
      hash = "sha256-vEZUBA+KeDHgqZvzrAN6ramZ5D4iqQdVU+qFOK/39co=";
    };

    vendorHash = "sha256-hYEXzKrACpSyvrAYbV0jkX504Ix/ch2PVrhksYKFhwE=";

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
