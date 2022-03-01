FROM kong:ubuntu as builder

USER root
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
      && apt-get -y install --no-install-recommends gcc \
                                                    patch \
                                                    libyaml-dev \
                                                    libssl-dev
USER kong

WORKDIR /goks/

CMD ["./gen-lua-tree.sh"]