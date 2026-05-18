FROM denoland/deno:alpine

WORKDIR /app

COPY deno.json deno.lock ./
RUN deno install --frozen
RUN deno clean

COPY . .

USER deno
CMD ["deno", "run", "--allow-net", "--allow-env", "--env-file", "--allow-read", "--allow-ffi" , "--allow-sys", "--allow-write=/tmp", "src/main.ts"]
