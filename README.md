filedist
========
demo版， 开发中 (developing)
智能文件分发工具 (包含一个控制中心，和客户端）

smart file distribution tool. (include a control center and clients)

## control center
The name is "filedist".

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

## history
- 2013-7-20 （没去成爬山，好遗憾啊） 客户端提供文件服务器和执行命令，两种基本功能
- 2013-7-21  noahdes auto deploy fire script, (mkpkg.sh setup.sh)
- 2013-7-22  net dial timeout, file server
- 2013-7-23  wget+md5sum function

## TODO
- client

	--kill
	--dir
	--password
	
	server specfiy url
	
	attach console
	
	more details of ps（time， args， user， dir ...)

- filedist

    wget result(md5sum check)
