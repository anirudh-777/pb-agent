# Contributing

1. Discuss substantial capability or policy changes in an issue first.
2. Create a focused branch and include tests with every behavior change.
3. Run:

   ```sh
   gofmt -w cmd internal
   go test -race ./...
   go vet ./...
   ```

4. Sign commits using the Developer Certificate of Origin:

   ```sh
   git commit -s
   ```

Policy weakening, new secret-bearing surfaces, raw request execution, SQL, and
remote transports require a threat-model update and maintainer approval.
