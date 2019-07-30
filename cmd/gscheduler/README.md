gscheduler
====

Shutdown instances with executing functinos in your console.

# Install

`go get -v github.com/future-architect/gcp-instance-scheduler/cmd/gscheduler`

# Usage

You can designate project id and timeout length by using flags.

```
gscheduler [flags]

Flags:
      --config string    config file (default is $HOME/.gscheduler.yaml)
  -h, --help             help for gscheduler
  -p, --project string   project id (defautl $GCP_PROJECT)
      --timeout string   set timeout seconds (default "60")
  -t, --toggle           Help message for toggle
```