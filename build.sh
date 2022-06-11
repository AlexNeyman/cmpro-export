build_dir="$(pwd)"/_build
mkdir -p "$build_dir"
go build -o "$build_dir" .
