# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
"""Client and server classes corresponding to protobuf-defined services."""
import grpc

import prediction_pb2 as prediction__pb2


class PredictionStub(object):
    """Missing associated documentation comment in .proto file."""

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.GetSubscriberPrediction = channel.unary_unary(
                '/prediction.Prediction/GetSubscriberPrediction',
                request_serializer=prediction__pb2.Message.SerializeToString,
                response_deserializer=prediction__pb2.MessageResponse.FromString,
                )
        self.GetReverseSubscriberPrediction = channel.unary_unary(
                '/prediction.Prediction/GetReverseSubscriberPrediction',
                request_serializer=prediction__pb2.Message.SerializeToString,
                response_deserializer=prediction__pb2.MessageResponse.FromString,
                )


class PredictionServicer(object):
    """Missing associated documentation comment in .proto file."""

    def GetSubscriberPrediction(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def GetReverseSubscriberPrediction(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_PredictionServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'GetSubscriberPrediction': grpc.unary_unary_rpc_method_handler(
                    servicer.GetSubscriberPrediction,
                    request_deserializer=prediction__pb2.Message.FromString,
                    response_serializer=prediction__pb2.MessageResponse.SerializeToString,
            ),
            'GetReverseSubscriberPrediction': grpc.unary_unary_rpc_method_handler(
                    servicer.GetReverseSubscriberPrediction,
                    request_deserializer=prediction__pb2.Message.FromString,
                    response_serializer=prediction__pb2.MessageResponse.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'prediction.Prediction', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))


 # This class is part of an EXPERIMENTAL API.
class Prediction(object):
    """Missing associated documentation comment in .proto file."""

    @staticmethod
    def GetSubscriberPrediction(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/prediction.Prediction/GetSubscriberPrediction',
            prediction__pb2.Message.SerializeToString,
            prediction__pb2.MessageResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def GetReverseSubscriberPrediction(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/prediction.Prediction/GetReverseSubscriberPrediction',
            prediction__pb2.Message.SerializeToString,
            prediction__pb2.MessageResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)
