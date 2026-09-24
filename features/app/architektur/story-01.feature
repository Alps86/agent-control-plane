# language: de
@story-01 @arc-01
Funktionalität: Go-Server über die öffentliche HTTP-Grenze erreichen
  Als Betreiber
  möchte ich den gestarteten Server über HTTP erreichen,
  damit das Anwendungsskelett von außen überprüfbar ist.

  Szenario: Der gestartete Server beantwortet eine HTTP-Smoke-Anfrage
    Angenommen der Go-Server ist auf einem freien lokalen Port gestartet
    Wenn ich über HTTP GET "/health" an den Server sende
    Dann antwortet der Server mit dem HTTP-Status 200
