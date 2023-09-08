from app.utility.base_world import BaseWorld

from plugins.ot.app.ot_gui import OTGUI
from plugins.ot.app.ot_api import OTAPI

name = 'OT'
description = 'The OT plugin for Caldera provides adversary emulation abilities specific to Operational Technology.'
address = '/plugin/ot/gui'
access = BaseWorld.Access.RED


async def enable(services):
    app    = services.get('app_svc').application
    ot_gui = OTGUI(services, name=name, description=description)

    app.router.add_static('/ot', 'plugins/ot/static/', append_version=True)
    app.router.add_route('GET', '/plugin/ot/gui', ot_gui.splash)

    ot_api = OTAPI(services)
    # Add API routes here
    app.router.add_route('POST', '/plugin/ot/mirror', ot_api.mirror)

