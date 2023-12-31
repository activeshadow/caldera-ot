import logging

from aiohttp_jinja2 import template

from app.service.auth_svc   import for_all_public_methods, check_authorization
from app.utility.base_world import BaseWorld

from plugins.ot.app.ot_svc import OTService

@for_all_public_methods(check_authorization)
class OTGUI(BaseWorld):
    def __init__(self, services, name, description):
        self.name        = name
        self.description = description
        self.services    = services
        self.modbus_svc  = OTService(services)

        self.data_svc = services.get('data_svc')
        self.auth_svc = services.get('auth_svc')

        self.log = logging.getLogger('ot_gui')

    @template('ot.html')
    async def splash(self, request):
        abilities = {
            a.ability_id: {
                "name"       :    a.name,
                "tactic"     :    a.tactic,
                "technique_id":   a.technique_id,
                "technique_name": a.technique_name,
                "description":    a.description.replace('\n', '<br>')  # nicer display
            }

            for a in await self.data_svc.locate('abilities')
            if await a.which_plugin() == 'ot'
        }

        abilities = list(abilities.values())

        return dict(name=self.name, description=self.description, abilities=abilities)
