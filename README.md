[![Build Status](https://travis-ci.org/shadyabhi/jqtop.png)](https://travis-ci.org/shadyabhi/jqtop)
[![codecov](https://codecov.io/gh/shadyabhi/jqtop/branch/master/graph/badge.svg)](https://codecov.io/gh/shadyabhi/jqtop)
[![Go Report Card](https://goreportcard.com/badge/github.com/shadyabhi/jqtop)](https://goreportcard.com/report/github.com/shadyabhi/jqtop)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/shadyabhi/jqtop)

   * [jqtop](#jqtop)
      * [What does it do?](#what-does-it-do)
      * [Short tutorial](#short-tutorial)
         * [Existing Fields](#existing-fields)
            * [Get stats for unique "requests" being made](#get-stats-for-unique-requests-being-made)
         * [Derived Fields](#derived-fields)
            * [Get stats on various HTTP methods being hit](#get-stats-on-various-http-methods-being-hit)
         * [Show stats for multiple fields](#show-stats-for-multiple-fields)
         * [Filtering](#filtering)
            * [Show top paths with http method as POST.](#show-top-paths-with-http-method-as-post)
            
# jqtop

jqtop is a tool to analyze json logs or equivalent stream of inputs in realtime.

## What does it do?

**For eg, for a stream of logs like**:-

```
{
  "time_local": "04/Oct/2018:06:30:27 +0000",
  "remote_addr": "69.162.124.230",
  "remote_user": "",
  "request": "HEAD /somepath.php HTTP/1.1",
  "status": "200",
  "body_bytes_sent": "0",
  "request_time": "0.000",
  "http_referrer": "https://abhi.host",
  "http_user_agent": "Mozilla/5.0+(compatible; UptimeRobot/2.0; http://www.uptimerobot.com/)"
}
```
**We can answer questions like**:-

* Stats for `request`.
* Stats for `derived fields` like `http_method` using `request` field. 
* Filter lines that should be processed for stats

**For the impatient folks, here's a sample usage:-**

* Creates a new `derived` field, `http_method` by applying regex on already existing field `request`. 
* Similarly, derive one more field `paths` which extracts path from the field `request`. 
* Filter only lines that have `http_method == "GET"`. 

![Alt Text](https://shadyabhi.keybase.pub/jqtop_demo_v1.gif)

## Short tutorial

The CLI tool takes 3 basic parameters:-

* `--file`: File that needs to be tailed.
* `--fields`: Fields that need to be aggregated.
* `--filters`: Filters used to filter lines that we want.
* ... and others

```
$ ./jqtop -h
Usage: jqtop --file FILE [--interval INTERVAL] [--maxresult MAXRESULT] [--verbose] [--clearscreen] [--fields FIELDS] [--filters FILTERS]

Options:
  --file FILE            Path to file that will be read
  --interval INTERVAL, -i INTERVAL
                         Interval at which stats are calculated [default: 1]
  --maxresult MAXRESULT, -m MAXRESULT
                         Max results to show [default: 10]
  --verbose, -v
  --clearscreen, -c      Clear screen each time stats are shown
  --fields FIELDS        Fields that need to shown for stats
  --filters FILTERS      Filters to filter lines that'll be processed
  --help, -h
```

### Existing Fields

#### Get stats for unique "requests" being made

```
$ ./jqtop -file ./logfile --fields request
2018/10/04 12:58:47 Seeked ./logfile - &{Offset:0 Whence:2}
✖ Parse error rate: 0
➤ request
└──    3: GET / HTTP/1.1
└──    2: GET /search/label/software%20access%20point?m=1 HTTP/1.1
└──    1: POST /1hou.php HTTP/1.1
└──    1: POST /miao.php HTTP/1.1
└──    1: POST /linuxse.php HTTP/1.1
└──    1: POST /tomcat.php HTTP/1.1
└──    1: POST /she.php HTTP/1.1
└──    1: POST /boots.php HTTP/1.1
└──    1: POST /qw.php HTTP/1.1
└──    1: POST /test.php HTTP/1.1
```

By default, stats are aggregated every second. You can change the internal using `-i` option.

### Derived Fields

In the above json docs, we can see that there is no HTTP method field but a full string is available under "request" field.

#### Get stats on various HTTP methods being hit

We'll use a simple regex to derive the field "http_method" from the "request" field.

```
$ ./jqtop -file ./logfile --fields 'http_method = regex_capture(request, "(.*?) ")'
2018/10/05 06:45:23 Seeked ./logfile - &{Offset:0 Whence:2}
✖ Parse error rate: 0
➤ http_method
└──   18: GET
└──    4: HEAD

```

### Show stats for multiple fields

We'll now show results for 2 fields, 1 being a regular field and another being a derived field.
We can specify multiple fields by just deelimiting them with a semi-colon.

Let's show only top 2 results.

```
$ ./jqtop -file ./logfile --fields 'http_method = regex_capture(request, "(.*?) "); request' -m 2
2018/10/05 06:46:15 Seeked ./logfile - &{Offset:0 Whence:2}
✖ Parse error rate: 0
➤ request
└──    7: GET / HTTP/1.1
└──    2: GET /apple-touch-icon-precomposed.png HTTP/1.1

➤ http_method
└──   18: GET
└──    2: HEAD
```

### Filtering

This is a complex example that uses both derived fields and filters.

#### Show top paths with http method as POST.

* Create fields

There is no field which has `context_path` as it's value, so we'll have to create a new derived field.

```
paths = regex_capture(request, "[A-Z]+? (.*?) ")
```

As we need to filter logs that contain only POST, we'll need to create a new field `http_method` and filter based on that.

Our new `field` argument would be:-

```
paths = regex_capture(request, "[A-Z]+? (.*?) "); http_method = regex_capture(request, "(.*?) ");
```

* Create filters

To filter based on `http_method`, out `---filters` argument would look like:-

```
equals(http_method, "POST")
```

* Final CLI options

```
$ ./jqtop -file ./logfile --fields 'paths = regex_capture(request, "[A-Z]+? (.*?) "); http_method = regex_capture(request, "(.*?) ");' ---filters 'equals(http_method, "POST")'
2018/10/05 07:07:38 Seeked ./logfile - &{Offset:0 Whence:2}
✖ Parse error rate: 0
➤ http_method
└──   18: POST

➤ paths
└──    1: /db__.init.php
└──    1: /wshell.php
└──    1: /xshell.php
└──    1: /qq.php
└──    1: /db_dataml.php
└──    1: /wc.php
└──    1: /xx.php
└──    1: /w.php
└──    1: /db_session.init.php
└──    1: /wp-admins.php
```
