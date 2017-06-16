package main

import (
    "log"
    _ "github.com/go-sql-driver/mysql"
    "database/sql"
    "github.com/streadway/amqp"
    "encoding/json"
    "encoding/csv"
    "encoding/base64"
    "os"
    "bytes"
    "github.com/sendgrid/sendgrid-go"
    "github.com/sendgrid/sendgrid-go/helpers/mail"
)

type Message struct {
    Email   string `json:"email"`
    SQL string `json:"sql"`
}

type Data struct {
    Id string
    Descricao string
    Sacado sql.NullString
    Telefone sql.NullString
    ValorBruto sql.NullString `valor_bruto`
    ValorPrevisto sql.NullString `valor_previsto`
    DataVencimento sql.NullString `data_vencimento`
    UnidadeId sql.NullString `unidade_id`
    UnidadeNome sql.NullString `unidade_nome`
}

func failOnError(err error, msg string) {
    if err != nil {
        log.Fatalf("%s: %s", msg, err)
    }
}

func GenerateCSV(users [][]string) []byte {

    b := &bytes.Buffer{} // creates IO Writer
    //https://gist.github.com/DavidVaini/6905911
    writer := csv.NewWriter(b)

    for _, value := range users {
        writer.Write(value)
    }

    writer.Flush()

    return b.Bytes()

}

func kitchenSink(atta []byte) []byte {

    m := mail.NewV3Mail()

    from := mail.NewEmail("Example User", "reports@rabbitmq.com")

    m.SetFrom(from)

    m.Subject = "Reports da fila"

    p := mail.NewPersonalization()

    to := mail.NewEmail("Example User", "lucassrod@gmail.com")

    p.AddTos(to)

    m.AddPersonalizations(p)

    plainTextContent := "relatorio analitico gerado contas a pagar e receber do academico"

    c := mail.NewContent("text/plain", plainTextContent)

    m.AddContent(c)

    htmlContent := "<strong>relatorio analitico gerado contas a pagar e receber do academico</strong>"

    c = mail.NewContent("text/html", htmlContent)

    m.AddContent(c)

    a := mail.NewAttachment()
    sEnc := base64.StdEncoding.EncodeToString([]byte(atta))
    a.SetContent(sEnc)
    a.SetType("text/csv")
    a.SetFilename("report.csv")
    a.SetDisposition("attachment")
    a.SetContentID("Balance Sheet")
    m.AddAttachment(a)

    m.AddCategories("Relatorios")
    m.AddCategories("Contas a pagar e a receber academico")

    trackingSettings := mail.NewTrackingSettings()

    clickTrackingSettings := mail.NewClickTrackingSetting()

    clickTrackingSettings.SetEnable(true)

    clickTrackingSettings.SetEnableText(true)

    trackingSettings.SetClickTracking(clickTrackingSettings)

    openTrackingSetting := mail.NewOpenTrackingSetting()

    openTrackingSetting.SetEnable(true)

    openTrackingSetting.SetSubstitutionTag("Optional tag to replace with the open image in the body of the message")

    trackingSettings.SetOpenTracking(openTrackingSetting)

    m.SetTrackingSettings(trackingSettings)

    return mail.GetRequestBody(m)
}

func send_mail(body []byte) {
    request := sendgrid.GetRequest(os.Getenv("SENDGRID_API_KEY"), "/v3/mail/send", "https://api.sendgrid.com")
    request.Method = "POST"

    request.Body = body
    response, err := sendgrid.API(request)
    if err != nil {
        log.Println(err)
    } else {
        log.Println(response.StatusCode)
        log.Println(response.Body)
        log.Println(response.Headers)
    }
}

func main() {
    conn, err := amqp.Dial("amqp://root:root@localhost:5672/")
    failOnError(err, "Failed to connect to RabbitMQ")
    defer conn.Close()

    ch, err := conn.Channel()
    failOnError(err, "Failed to open a channel")
    defer ch.Close()

    q, err := ch.QueueDeclare(
        "contas_pagar_receber_academico", // name
        true,   // durable
        false,   // delete when unused
        false,   // exclusive
        false,   // no-wait
        nil,     // arguments
    )
    failOnError(err, "Failed to declare a queue")

    msgs, err := ch.Consume(
        q.Name, // queue
        "",     // consumer
        false,   // auto-ack
        false,  // exclusive
        false,  // no-local
        false,  // no-wait
        nil,    // args
    )
    failOnError(err, "Failed to register a consumer")

    db, err := sql.Open("mysql", "root:root@/root")
    failOnError(err, "Falha ao conectar ao mysql")
    defer db.Close()
    forever := make(chan bool)

    go func() {
        for d := range msgs {
            log.Printf("Received a message: %s", d.Body)

            msg := Message{}
            json.Unmarshal([]byte(d.Body), &msg)
            // log.Printf(msg)
            log.Printf(msg.SQL)

            sql := "SELECT r.id, r.descricao, r.sacado, r.telefone, r.valor_bruto, r.valor_previsto, DATE_FORMAT(r.data_vencimento, '%m-%d-%Y') AS data_vencimento, r.unidade_id, r.unidade_nome FROM relatorio_contas_pagar_receber AS r WHERE r.tipo = 'd' AND r.status = 'r' AND r.unidade_id IN (1) AND data_vencimento >= '2017-02-01 00:00:00' AND data_vencimento <= '2017-06-14 23:59:59' ORDER BY r.unidade_id, data_vencimento, r.sacado"

            rows, err := db.Query(sql)
            failOnError(err, "Erro ao executar sql")
            defer rows.Close()

            users := [][]string{}
            for rows.Next() {
                var r Data
                var lista []string
                err := rows.Scan(&r.Id, &r.Descricao, &r.Sacado, &r.Telefone, &r.ValorBruto, &r.ValorPrevisto, &r.DataVencimento, &r.UnidadeId, &r.UnidadeNome)
                failOnError(err, "erro ao scanear camps")
                if r.Sacado.Valid {
                    log.Printf("Valido")
                    log.Printf(r.Sacado.String)
                } else {
                    log.Printf("Nulo")
                }
                if r.ValorPrevisto.Valid {
                    log.Printf("Valido")
                    log.Printf(r.ValorPrevisto.String)
                } else {
                    log.Printf("Nulo")
                }
                log.Printf("-----------------------------------")
                lista = append(lista, r.Id)
                lista = append(lista, r.Descricao)
                lista = append(lista, r.Sacado.String)
                lista = append(lista, r.Telefone.String)
                lista = append(lista, r.ValorBruto.String)
                lista = append(lista, r.ValorPrevisto.String)
                lista = append(lista, r.DataVencimento.String)
                lista = append(lista, r.UnidadeId.String)
                lista = append(lista, r.UnidadeNome.String)
                users = append(users, lista)
            }

            attach := GenerateCSV(users)

            email := kitchenSink(attach)

            send_mail(email)

            d.Ack(false)
        }
    }()

    log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
    <-forever
}