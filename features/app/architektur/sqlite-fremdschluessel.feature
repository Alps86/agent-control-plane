# language: de
@sqlite @fremdschluessel
Funktionalität: Fremdschlüssel im lokalen SQLite-Bestand durchsetzen
  Als Betreiber
  möchte ich ungültige Referenzen beim Schreiben und Öffnen abweisen,
  damit kein widersprüchlicher Bestand unbemerkt weiterläuft.

  Szenario: Eine Migration mit ungültiger Referenz scheitert auch nach erneutem Öffnen
    Angenommen ein lokaler SQLite-Bestand mit verknüpften Tabellen ist angelegt
    Wenn eine Migration einen Kinddatensatz ohne Elternsatz schreiben will
    Dann wird der Fremdschlüsselverstoß abgewiesen und die Schemaversion bleibt unverändert
    Wenn ich denselben Bestand erneut öffne und den Fremdschlüsselverstoß wiederhole
    Dann wird der Fremdschlüsselverstoß abgewiesen und die Schemaversion bleibt unverändert

  Szenario: Ein vorhandener Bestand mit ungültiger Referenz verhindert den Start
    Angenommen ein vorhandener SQLite-Bestand enthält eine ungültige Fremdschlüsselreferenz
    Wenn ich den Bestand über die öffentliche SQLite-Grenze öffne
    Dann wird der beschädigte Bestand mit einem verständlichen Fremdschlüsselfehler abgewiesen
