ARG RUNTIME_TAG=latest
FROM filvenus/venus-runtime:${RUNTIME_TAG}

ARG BUILD_TARGET=venus-tool
ENV VENUS_COMPONENT=${BUILD_TARGET}

# copy the app from build env
COPY ./${BUILD_TARGET} /app/${BUILD_TARGET}
COPY ./dashboard/build /app/dashboard/build

ENTRYPOINT ["/script/init.sh"]
