FROM alpine:latest

WORKDIR /

ENV KUBECTL_VERSION=1.25.13

ENV YQ_VERSION=4.33.3

# RUN apk add --update -t deps curl openssl

RUN apk add --update -t deps openssl

RUN wget https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl -O /usr/local/bin/kubectl \
  && chmod +x /usr/local/bin/kubectl

# RUN curl -L https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl \
#   && chmod +x /usr/local/bin/kubectl

# RUN wget https://github.com/mikefarah/yq/releases/download/v${YQ_VERSION}/yq_linux_amd64 -o /usr/local/bin/yq \
#   && chmod +x /usr/local/bin/yq

COPY scripts/ .

ENTRYPOINT ["sh", "generate-pki.sh"]