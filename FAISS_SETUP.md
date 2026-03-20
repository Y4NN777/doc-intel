# FAISS Setup Instructions

This project uses FAISS for vector similarity search via CGO bindings.

## Prerequisites

### Install FAISS C Library

```bash
# Clone FAISS
git clone https://github.com/facebookresearch/faiss.git
cd faiss

# Build with C API enabled
cmake -B build \
  -DFAISS_ENABLE_GPU=OFF \
  -DFAISS_ENABLE_C_API=ON \
  -DBUILD_SHARED_LIBS=ON \
  .

make -C build -j
sudo make -C build install

# Install shared library
sudo cp build/c_api/libfaiss_c.so /usr/lib/
sudo ldconfig
```

### Install Go FAISS Bindings

```bash
go get github.com/DataIntelligenceCrew/go-faiss
```

## Building

```bash
make build
```

## Testing

```bash
make test
```
