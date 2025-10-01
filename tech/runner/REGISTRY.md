# Registry

- [Docker hub](https://hub.docker.com/) is the most popular container registry, but many companies use private registries for security and control.
- Self Certificate: If you use a self-signed certificate, you need to add the certificate to your Docker daemon's trusted certificates.
- Private Registry: You can set up your own private Docker registry using the `Docker Registry` image or use third-party services like `Harbor`, `JFrog Artifactory`, or `AWS ECR`.

## [Docker Hub](https://hub.docker.com/)

- Login to Docker Hub:

```bash
docker login
```

- Create tag for your image:

```bash
docker tag <image_id> <username>/<repository>:<tag>
```

- Push image to Docker Hub:

```bash
docker push <username>/<repository>:<tag>
```

- Pull image from Docker Hub:

```bash
docker pull <username>/<repository>:<tag>
```

## Private Registry

- Use `docker-compose` to set up a private registry with `Docker Registry`.

```yaml
services:
  registry:
    image: registry:2
    ports:
      - "5000:5000"
    environment:
      REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY: /var/lib/registry
      REGISTRY_HTTP_TLS_CERTIFICATE: /certs/domain.crt
      REGISTRY_HTTP_TLS_KEY: /certs/domain.key
    volumes:
      - ./data:/var/lib/registry
      - ./certs:/certs
```

- Trying connect to registry:

```bash
curl https://localhost:5000/v2/_catalog
```

- Configure Docker to trust the self-signed certificate:

```bash
mkdir /etc/docker/certs.d/localhost:5000
cp /certs/domain.crt /etc/docker/certs.d/localhost:5000/ca.crt
systemctl restart docker
```

## Harbor Registry

- Should have `VPS` or `Server` to install Harbor registry. You can use `AWS`, `DigitalOcean`, `Linode`, `Vultr`, etc.
- Install [`Harbor` registry](https://goharbor.io/docs/2.0.0/install-config/)
