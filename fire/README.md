## client
The name is "fire". It offer some basic services(execute command, file server)

### usage
    fire [OPTIONS] args ...

    Help Options:
      -h, --help=        Show this help message

    Type:
      -m, --type=        type [run|ps|wait] (run)

    Run:
      -H, --host=        host to connect (localhost)
      -t, --timeout=     time out [s|m|h] (0s)
      -b, --background   run in background
      -e, --env=         add env to runner,multi support. eg -e PATH=/bin -e TMPDIR=/tmp
          --dialtimeout= dial timeout,unit seconds (2s)

    Serve:
      -d, --daemon       run as server (false)
      -p, --port=        port to connect or serve (4456)
          --fs=          open a http file server (/tmp/:/home/)
          
### example
    JOB_ID=`fire -H example1.com -t 2s -b sh -c "echo start; sleep 5s; echo done"`
    fire -m wait $JOB_ID
    # exepect output, and exit code is 128(because it runs timeout)
    #
    # start
    #
