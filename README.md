# fluentd_chunk_parser
Fluentd secondary dump file reprocesser

This program can resend failed chunk file on fluentd(secondary_file)

Send to Fluentd HTTP input / File Output(json) / stdout

1. Build

   ```sh
   git clone https://github.com/nsheo/fluentd_chunk_parser.git
   cd fluentd_chunk_parser
   go get
   go build
   ```

2. Setting

   ```json
   {
     "settings": [
       {
         "file_name": "test.bin", //target file to read(req)
         "base64_decode_target": ["host", "ident", "message", "point", "worker_id"], //log string text that need base64 decoding(req)
         "send_target": "", //send target(stdout,file,fluentd) - default stdout
         "target_path": "" //send path(file/fluentd)
       }
     ]
   }
   ```

   

3. Run

   ```shell
   ./fluentd_chunk_parser -settings=./settings.json
   ```

   