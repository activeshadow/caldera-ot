import logging


class OTService:
    def __init__(self, services):
        self.services = services
        self.file_svc = services.get('file_svc')

        self.log = logging.getLogger('ot_svc')

    async def foo(self):
        return 'bar'
