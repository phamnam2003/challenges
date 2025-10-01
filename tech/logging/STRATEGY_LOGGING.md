# Analysis Logging Strategy

- From `logs` file -> analyze logs -> recommend plan collect and format logs -> apply standard rules into file configuration
- Logging file can be `text`, `json`, `xml`, `csv` .... You must rewrite logs to `json` format for easy to parse and analyze.

## TEXT LOG LINE

- This format logs often hard to parse and analyze, because each line can be different format, so you need to write complex regex to parse log line. It often use in `Java Spring Boot` application, `morgan` library in `NodeJS`.

```stdout
# Spring Boot default log format
2025-09-18T11:18:46.890Z  INFO 139141 --- [myapp] [           main] o.s.b.d.f.logexample.MyApplication       : Starting MyApplication using Java 17.0.16 with PID 139141 (/opt/apps/myapp.jar started by myuser in /opt/apps/)
2025-09-18T11:18:46.901Z  INFO 139141 --- [myapp] [           main] o.s.b.d.f.logexample.MyApplication       : No active profile set, falling back to 1 default profile: "default"
2025-09-18T11:18:50.607Z  INFO 139141 --- [myapp] [           main] o.s.b.w.embedded.tomcat.TomcatWebServer  : Tomcat initialized with port 8080 (http)
2025-09-18T11:18:50.674Z  INFO 139141 --- [myapp] [           main] o.apache.catalina.core.StandardService   : Starting service [Tomcat]
2025-09-18T11:18:50.677Z  INFO 139141 --- [myapp] [           main] o.apache.catalina.core.StandardEngine    : Starting Servlet engine: [Apache Tomcat/10.1.46]
2025-09-18T11:18:50.863Z  INFO 139141 --- [myapp] [           main] o.a.c.c.C.[Tomcat].[localhost].[/]       : Initializing Spring embedded WebApplicationContext
2025-09-18T11:18:50.873Z  INFO 139141 --- [myapp] [           main] w.s.c.ServletWebServerApplicationContext : Root WebApplicationContext: initialization completed in 3724 ms
2025-09-18T11:18:52.159Z  INFO 139141 --- [myapp] [           main] o.s.b.w.embedded.tomcat.TomcatWebServer  : Tomcat started on port 8080 (http) with context path '/'
2025-09-18T11:18:52.208Z  INFO 139141 --- [myapp] [           main] o.s.b.d.f.logexample.MyApplication       : Started MyApplication in 7.341 seconds (process running for 8.401)
2025-09-18T11:18:52.230Z  INFO 139141 --- [myapp] [ionShutdownHook] o.s.b.w.e.tomcat.GracefulShutdown        : Commencing graceful shutdown. Waiting for active requests to complete
2025-09-18T11:18:52.249Z  INFO 139141 --- [myapp] [tomcat-shutdown] o.s.b.w.e.tomcat.GracefulShutdown        : Graceful shutdown complete

# NodeJS morgan log format
::1 - - [27/Nov/2024:06:21:42 +0000] "GET /combined HTTP/1.1" 200 2 "-" "curl/8.7.1"
::1 - - [27/Nov/2024:06:21:46 +0000] "GET /common HTTP/1.1" 200 2
::1 - - [02/Jun/2025:10:15:14 +0000] "GET / HTTP/1.1" 304 - "-" "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36"
::1 - - [02/Jun/2025:10:15:19 +0000] "GET /about HTTP/1.1" 304 - "-" "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36"
```

- Parser: You need to write complex regex to parse log line, it can be hard to maintain and error-prone because `Elasticsearch` use logs with `JSON` format.

```json
{
  "timestamp": "2025-09-18T11:18:46.890Z",
  "log": { "level": "info" },
  "process": { "pid": 139141 },
  "lables": { "thread": "myapp", "user": "myuser", "env": "prod" },
  "thread": "main",
  "logger": "o.s.b.d.f.logexample.MyApplication",
  "message": "Starting MyApplication using Java 17.0.16 with PID 139141 (/opt/apps/myapp.jar started by myuser in /opt/apps/)",
  "service": "myapp",
  "event": { "dataset": "springboot.log" }
}

{
  "timestamp": "2024-11-27T06:21:42+00:00",
  "log": { "level": "info" },
  "lables": { "host": "::1", "env": "dev" },
  "message": "Hibernate: SELECT * FROM accounts where id = 1 LIMIT 1",
  "sql": {
    "query": "SELECT * FROM accounts where id = 1 LIMIT 1;",
    "status": "success",
    "rows": 1,
    "duration_ms": 12
  },
  "event": { "dataset": "nodejs.morgan" }
}
```

- Analyzed to create `index pattern` in `Kibana` for easy to search and visualize logs by field. Logs with low throughput, much index or the days have less logs -> server shared logs with expensive overhead.
