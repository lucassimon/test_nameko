from flask import Flask, request
from flasgger import Swagger
from nameko.standalone.rpc import ClusterRpcProxy

app = Flask(__name__)
Swagger(app)
CONFIG = {'AMQP_URI': "amqp://root:root@localhost"}


@app.route('/compute', methods=['POST'])
def compute():
    """
    Micro Service Based Compute and Mail API
    This API is made with Flask, Flasgger and Nameko
    ---
    parameters:
      - name: body
        in: body
        required: true
        schema:
          id: data
          properties:
            operation:
              type: string
              enum:
                - sum
                - mul
                - sub
                - div
            email:
              type: string
            value:
              type: integer
            other:
              type: integer
    responses:
      200:
        description: Please wait the calculation, you'll receive an email with results
    """

    operation = request.json.get('operation')

    value = request.json.get('value')

    other = request.json.get('other')

    email = request.json.get('email')

    msg = "Please wait the calculation, you'll receive an email with results"

    subject = "API Notification"

    with ClusterRpcProxy(CONFIG) as rpc:

        # asynchronously spawning and email notification

        hello_res = rpc.mail.send.call_async(email, subject, msg)

        hello_res.result()  # "hello-x-y"

        # asynchronously spawning the compute task

        world_res = rpc.compute.compute.call_async(
            operation, value, other, email
        )

        world_res.result()  # "world-x-y"
        return msg, 200


app.run(debug=True)