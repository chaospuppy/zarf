FROM alpine:latest

WORKDIR /binary
RUN apk add gcc musl-dev && \
    wget -O darkhttpd.c https://raw.githubusercontent.com/emikulic/darkhttpd/master/darkhttpd.c && \
    cc -static -Os -o darkhttpd darkhttpd.c

WORKDIR /site
RUN wget https://js-dos.com/6.22/current/js-dos.js && \
    wget https://js-dos.com/6.22/current/wdosbox.js && \
    wget https://js-dos.com/6.22/current/wdosbox.wasm.js

RUN wget -O aladdin.zip "https://web.archive.org/web/20190303222445if_/https://www.dosgames.com/files/DOSBOX_ALADDIN.ZIP"
RUN wget -O doom.zip "https://archive.org/download/DoomsharewareEpisode/doom.ZIP"
RUN wget -O quake.zip "https://web.archive.org/web/20190303223506if_/https://www.dosgames.com/files/DOSBOX_QUAKE.ZIP"
RUN wget -O warcraft-ii.zip "https://web.archive.org/web/20190303222732if_/https://www.dosgames.com/files/DOSBOX_WAR2.ZIP"

RUN wget -O aladdin.png "https://image.dosgamesarchive.com/screenshots/aladdem-4.png" && \
    wget -O doom.png "https://image.dosgamesarchive.com/screenshots/doom01.png" && \
    wget -O quake.png "https://image.dosgamesarchive.com/screenshots/quake13.png" && \
    wget -O warcraft-ii.png "https://image.dosgamesarchive.com/screenshots/war2demo3.png"


COPY index.html .

FROM scratch
COPY --from=0 /site /site
COPY --from=0 /binary /binary

WORKDIR /site
ENTRYPOINT ["/binary/darkhttpd", "/site", "--port", "8000"]

# docker buildx build --push --platform linux/arm/v7,linux/arm64/v8,linux/amd64 --tag defenseunicorns/zarf-game:multi-tile .
