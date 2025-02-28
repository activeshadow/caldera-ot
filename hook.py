from app.utility.base_world import BaseWorld

from plugins.ot.app.ot_svc import OTService

name = 'OT'
description = 'The OT plugin for Caldera provides adversary emulation abilities specific to Operational Technology.'
address = '/plugin/ot/gui'
access = BaseWorld.Access.RED


async def enable(services):
    ot_svc = OTService(services, name, description)
    app    = services.get('app_svc').application

    app.router.add_route('GET', '/plugin/ot/gui',  ot_svc.splash)
    app.router.add_route('GET', '/plugin/ot/data', ot_svc.plugin_data)
