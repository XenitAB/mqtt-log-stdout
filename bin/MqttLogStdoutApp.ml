let id x = x

let subscribe () =
  (* Get the MQTT port from environment *)
  let port =
    try Sys.getenv "MQTT_PORT" |> int_of_string
    with _ -> failwith "MQTT_PORT not correctly set in environment"
  in
  (* Get the MQTT hosts from environment *)
  let hosts =
    [
      Sys.getenv_opt "MQTT_HOST_1";
      Sys.getenv_opt "MQTT_HOST_2";
      Sys.getenv_opt "MQTT_HOST_3";
    ]
    |> List.filter_map id
  in
  (* Make sure there are at least 1 MQTT host *)
  let () =
    if List.length hosts = 0 then
      failwith "Must set at least one of MQTT_HOST_{1,2,3} in environment"
  in
  (* Get the MQTT topic we should listen on from environment *)
  let topic =
    try Sys.getenv "LOG_TOPIC"
    with _ -> failwith "Must set LOG_TOPIC in environment"
  in
  (* Create a client_id with some randomness at the end to not clash *)
  let client_id =
    Random.self_init ();
    let random_suffix =
      String.init 5 (fun _ -> Char.chr (Random.int 57 + 65))
    in
    Printf.sprintf "mqtt-log-stdout-%s" random_suffix
  in
  (* Open Lwt.Syntax to be able to use let*/let+ *)
  let open Lwt.Syntax in
  (* Create a MQTT client and connect it *)
  let* client = Mqtt_client.connect ~keep_alive:60 ~id:client_id ~port hosts in
  (* Start subscribing to the topic we want *)
  let* () =
    Mqtt_client.subscribe [ (topic, Mqtt_client.Atleast_once) ] client
  in
  (* Get a stream of messages from the topic *)
  let stream = Mqtt_client.messages client in
  (* For each message on the stream we print it to stdout *)
  Lwt_stream.iter (fun (_, message) -> print_endline message) stream

let () = Lwt_main.run (subscribe ())
