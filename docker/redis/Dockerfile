FROM bitnami/redis:latest

ARG REDIS_PASSWORD=V3nomous
ARG CONF_PATH=/usr/local/etc/redis/redis.conf

WORKDIR /

CMD ["redis-server", "/usr/local/etc/redis/redis.conf"]