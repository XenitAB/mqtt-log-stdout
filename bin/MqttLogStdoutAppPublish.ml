module C = Mqtt_client

let host = "127.0.0.1"

let port = 1883

let pub_example () =
  let open Lwt.Syntax in
  let* () = Lwt_io.printl "Starting publisher..." in
  let* client = C.connect ~id:"client-1" ~port [ host ] in
  let rec loop () =
    let* () = Lwt_io.printl "Publishing..." in
    let* line = Lwt_io.read_line Lwt_io.stdin in
    let* () =
      C.publish ~qos:C.Atleast_once ~topic:"xotclient/x/v1/log" line client
    in
    let* () = Lwt_io.printl "Published." in
    loop ()
  in
  loop ()

let () = Lwt_main.run (pub_example ())
