# Build the image
docker build -t sailor-console .

# Run the container with a volume mount for persistent database storage
docker run -d \
  --name sailor-server \
  -p 7766:7766 \
  -v /path/to/your/configs:/app/configs \
  sailor-console