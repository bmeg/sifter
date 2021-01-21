#!/usr/bin/env python

import sys
import grpc
import time
import json
import logging

import traceback
import exec_pb2
import exec_pb2_grpc
from concurrent import futures

from inspect import currentframe, getframeinfo

_ONE_DAY_IN_SECONDS = 60 * 60 * 24

#logging.basicConfig(filename='sifter-exec.log', level=logging.INFO)

class CallableCode:
    def __init__(self, funcName, code):
        self.name = funcName
        self.code = code
        self.env = {}
        exec(code, self.env)

    def call(self, values):
        return self.env[self.name](*values)


class PySifterExec:

    def __init__(self):
        self.code = {}
        self.code_num = 0
        self.source = {}

    def Compile(self, request, context):
        logging.info("Compile: %s", request.code)
        c = CallableCode(request.function, request.code)
        self.code[self.code_num] = c
        self.source[self.code_num] = request.code
        out = exec_pb2.CompileResult()
        out.id = self.code_num
        self.code_num += 1
        return out

    def Call(self, request, context):
        c = self.code[request.code]
        frameinfo = getframeinfo(currentframe())
        try:
            logging.info("calling %s on %s" % (request.code, request.data))
            data = json.loads(request.data)
            value = c.env[c.name](*data) #c.call(data)
            o = exec_pb2.Result()
            o.data = json.dumps(value)
            return o
        except Exception as e:
            o = exec_pb2.Result()
            ln = sys.exc_info()[2].tb_next.tb_lineno - 1
            code = self.source[request.code].split("\n")
            errorLine = ""
            if ln < len(code):
                errorLine = code[ln]
            o.error = traceback.format_exc() + "\n" + errorLine
            logging.info("ExecError: %s" % (o.error))
            return o

    def Process(self, request_iterator, context):
        logging.info("Calling Processor")
        for req in request_iterator:
            if req.code in self.code:
                c = self.code[req.code]
                try:
                    logging.info("calling %s on %s" % (req.code, req.data))
                    data = json.loads(req.data)
                    value = c.call(data)
                    o = exec_pb2.Result()
                    o.data = json.dumps(value)
                    yield o
                except Exception as e:
                    o = exec_pb2.Result()
                    o.error = str(e)
                    logging.info("ExecError: %s" % (o.error))
                    yield o



if __name__ == "__main__":

    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    exec_pb2_grpc.add_ExecutorServicer_to_server(
      PySifterExec(), server)
    port = 50000
    while True:
        new_port = server.add_insecure_port('[::]:%s' % port)
        if new_port != 0:
            break
        port += 1
    port = new_port
    server.start()
    print(port, flush=True)
    logging.info("Server started on port %d" % (port))

    try:
        while True:
            time.sleep(_ONE_DAY_IN_SECONDS)
    except KeyboardInterrupt:
        server.stop(0)
