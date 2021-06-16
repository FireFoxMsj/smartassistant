this_dir = File.expand_path(File.dirname(__FILE__))
$LOAD_PATH.unshift(this_dir) unless $LOAD_PATH.include?(this_dir)

require "plugin_pb"
require "plugin_services_pb"
require "grpc"

def main
  puts "incoming"
  stub = Comet::Comet::Stub.new('localhost:9234', :this_channel_is_insecure)
  st = stub.bid_stream(Comet::Request.new(body: nil), {metadata: {plugin_id: "value1"}})
  begin
     puts("在循环语句中 i = #$i" )
  end while true
end
main
