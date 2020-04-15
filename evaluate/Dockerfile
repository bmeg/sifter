FROM python

RUN pip install grpcio-tools
ADD sifter-exec.py exec_pb2.py exec_pb2_grpc.py /opt/
EXPOSE 50000

ENTRYPOINT /opt/sifter-exec.py
