FROM ubuntu:18.04
LABEL maintainer="Alexey Sudachen <alexey@sudachen.name>"

USER root
WORKDIR "/"
COPY .bin/go-spacemesh /local-testnet/bin/
COPY .bin/poet /local-testnet/bin/
COPY .bin/local-testnet /local-testnet/bin/
COPY genesis_accounts.json /local-testnet/
CMD ["/local-testnet/bin/local-testnet"]
EXPOSE 19090 19091 19190 19191

