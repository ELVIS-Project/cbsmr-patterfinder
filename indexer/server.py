# :todo fix paths to proto/
import os
import sys
sys.path.append(os.path.join(os.path.abspath(os.path.dirname(__file__)), os.pardir))
sys.path.append(os.path.join(os.path.abspath(os.path.dirname(__file__)), os.pardir, 'proto'))

from concurrent import futures
import time
import logging
import base64

import grpc

from indexer import indexers
from proto import smr_pb2, smr_pb2_grpc

_ONE_DAY_IN_SECONDS = 60 * 60 * 24

OPTIONS = [
        ('grpc.max_send_message_length', int(os.environ['INDEXER_MSG_SIZE'])),
        ('grpc.max_receive_message_length', int(os.environ['INDEXER_MSG_SIZE']))]

class Index(smr_pb2_grpc.IndexServicer):

    def _handle_symbolic_data(self, request):
        if request.encoding == smr_pb2.IndexRequest.BASE64:
            print("handling BASE64")
            
            sd = base64.b64decode(request.symbolic_data)
            return sd
        elif request.encoding == smr_pb2.IndexRequest.UTF8:
            print("handling UTF8")
            return request.symbolic_data.decode('utf-8')
        else:    
            print("Handling else")
            return request.symbolic_data

    def IndexNotes(self, request, context):
        sd = self._handle_symbolic_data(request)

        notes = indexers.notes(sd)[['onset', 'offset', 'pitch-chr']]
        pb_notes = smr_pb2.IndexResponse(notes = [
            smr_pb2.Note(
                onset=on,
                offset=off,
                pitch=int(p),
                piece_idx=idx)
                for idx, (_, on, off, p) in enumerate(notes.itertuples())])

        pb_notes.notes.sort(key=lambda n: (n.piece_idx))
        return pb_notes

def new_server(port):
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10), options = OPTIONS)
    smr_pb2_grpc.add_IndexServicer_to_server(Index(), server)
    server.add_insecure_port('[::]:' + port)
    return server

def serve(port):
    server = new_server(port)
    server.start()
    try:
        while True:
            time.sleep(_ONE_DAY_IN_SECONDS)
    except KeyboardInterrupt:
        server.stop(0)


if __name__ == '__main__':
    logging.basicConfig()
    logger = logging.getLogger('indexer')

    port = os.environ['INDEXER_PORT']

    serve(port)
