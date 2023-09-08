import re

from app.utility.base_parser import BaseParser

from app.objects.secondclass.c_fact         import Fact
from app.objects.secondclass.c_relationship import Relationship


class Parser(BaseParser):

    def parse(self, blob):
        relationships = []

        for line in self.line(blob):
            match = re.match(
                r'(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}) port (\d{1,5}) is open!',
                line,
            )

            if match:
                for mp in self.mappers:
                    if mp.source == 'modbus.server.ip' and match.group(2) == '502':
                        relationships.append(
                            Relationship(
                                source=Fact(mp.source, match.group(1)),
                                edge='isModbus',
                                target=Fact('modbus.server.port', match.group(2)),
                            )
                        )

                    if mp.source == 'dnp3.server.ip' and match.group(2) == '20000':
                        relationships.append(
                            Relationship(
                                source=Fact(mp.source, match.group(1)),
                                edge='isDNP3',
                                target=Fact('dnp3.server.port', match.group(2)),
                            )
                        )

        return relationships
