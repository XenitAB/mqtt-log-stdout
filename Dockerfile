FROM reasonnative/ocaml:4.10.1 as builder

RUN mkdir /app
WORKDIR /app

COPY package.json esy.lock mqtt-log-stdout.opam dune-project /app/

RUN esy install
RUN esy build-dependencies --release

COPY ./bin /app/bin

RUN esy dune build --profile=docker --release

RUN esy mv "#{self.target_dir / 'default' / 'bin' / 'MqttLogStdoutApp.exe'}" main.exe

RUN strip main.exe

FROM scratch as runtime

WORKDIR /app

COPY --from=builder /app/main.exe main.exe

ENTRYPOINT ["/app/main.exe"]
