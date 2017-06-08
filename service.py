import yagmail
from nameko.rpc import rpc, RpcProxy
import sendgrid
import base64
import os

from sendgrid.helpers.mail import *
try:
    # Python 3
    import urllib.request as urllib
except ImportError:
    # Python 2
    import urllib2 as urllib


class Mail(object):
    name = "mail"

    @rpc
    def send(self, to, subject, contents):

        sg = sendgrid.SendGridAPIClient(
            apikey=os.environ.get('SENDGRID_API_KEY')
        )

        file_path = 'file_path.txt'
        with open(file_path, 'rb') as f:
            arq = f.read()
            f.close()
        encoded = base64.b64encode(arq).decode()

        data = {
            "personalizations": [
                {
                    "to": [
                        {
                            "email": to
                        }
                    ],
                    "substitutions": {
                        "-name-": "Example User",
                        "-city-": "Denver"
                    },
                    "subject": subject
                },
            ],
            "from": {
                "email": "nameko@teste.com"
            },
            "content": [
                {
                    "type": "text/html",
                    "value": contents
                }
            ],
            # "attachments": [
            #     {
            #         "content": encoded,
            #         "filename": file_path
            #     }
            # ]

        }

        try:
            response = sg.client.mail.send.post(request_body=data)
            return response.status_code
        except Exception as e:
            print(e)
        except urllib.HTTPError as e:
            print (e)

        return 500


class Compute(object):
    name = "compute"
    mail = RpcProxy('mail')

    @rpc
    def compute(self, operation, value, other, email):
        operations = {
            'sum': lambda x, y: int(x) + int(y),
            'mul': lambda x, y: int(x) * int(y),
            'div': lambda x, y: int(x) / int(y),
            'sub': lambda x, y: int(x) - int(y)
        }

        try:
            result = operations[operation](value, other)
        except Exception as e:
            self.mail.send(email, "An error occurred", str(e))
            raise
        else:
            self.mail.send(
                email,
                "Your operation is complete!",
                "The result is: %s" % result
            )
            return result