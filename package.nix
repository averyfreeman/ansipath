{
  lib,
  python3Packages,
}:

python3Packages.buildPythonApplication rec {
  pname = "ansipath";
  version = "0.1.1";
  pyproject = true;

  src = ./.;

  nativeBuildInputs = [ python3Packages.hatchling ];

  pythonImportsCheck = [ "ansipath" ];

  meta = {
    description = "A colorized diagnostic view of shell PATH entries";
    homepage = "https://github.com/averyfreeman/ansipath";
    license = lib.licenses.gpl3Only;
    mainProgram = "ansipath";
    platforms = lib.platforms.unix;
  };
}
