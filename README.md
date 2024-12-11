# generator

This repository contains experimental code for SDK generation.

To run the generator:

```
# Build dotnet generator's docker
git clone git@github.com:googleapis/google-cloud-dotnet.git
cd google-cloud-dotnet
docker build -f Dockerfile.generator -t picard .

# Run the generator
go run ./cmd/generator generate -language dotnet
```

## License

Apache 2.0 - See [LICENSE] for more information.

[contributing]: CONTRIBUTING.md
[license]: LICENSE
