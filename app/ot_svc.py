import logging

from aiohttp        import web
from aiohttp_jinja2 import template


class OTService:
    def __init__(self, services, name, description):
        self.services    = services
        self.name        = name
        self.description = description

        self.data_svc = services.get('data_svc')
        self.log      = logging.getLogger('ot_svc')


    @template('ot.html')
    async def splash(self, request):
        data = await self._get_plugin_data()
        return data


    async def plugin_data(self, request):
        data = await self._get_plugin_data()
        return web.json_response(data)


    async def _get_plugin_data(self):
        abilities = {
            a.ability_id: {
                "name"           : a.name,
                "tactic"         : a.tactic,
                "technique_id"   : a.technique_id,
                "technique_name" : a.technique_name,
                "description"    : a.description.replace('\n', '<br>')
            }
            for a in await self.data_svc.locate('abilities')
            if await a.which_plugin() == 'ot'
        }

        abilities = list(abilities.values())
        return dict(name=self.name, description=self.description, abilities=abilities)
