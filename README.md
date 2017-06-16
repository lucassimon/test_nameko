Linha 3635
financeiro/reports.py

import pika
import json

	credentials = pika.PlainCredentials('root', 'root')
parameters = pika.ConnectionParameters(credentials=credentials)
	connection = pika.BlockingConnection(
			parameters

			)
channel = connection.channel()

	channel.queue_declare(
			queue="contas_pagar_receber_academico",
			durable=True,
			exclusive=False,
			auto_delete=False

			)

	data = {
		'email': 'lucassrod@gmail.com',
		'sql': '{}'.format(titulos.query)

	}

properties = pika.BasicProperties(
		app_id='cotemar',
		content_type='application/json',

		)

channel.basic_publish(
		exchange='',
		routing_key='contas_pagar_receber_academico',
		body=json.dumps(data),
		properties=properties

		)
