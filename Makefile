boot-consul:
	@ docker run -d -p "8500:8500" -p "8600:8600" --name=discovery hashicorp/consul agent -server -ui -node=server-1  -bootstrap-expect=1 -client="0.0.0.0"

discovery:
	@ docker run -d -p '8500:8500' -p 8600:8600/udp hashicorp/consul:latest agent -server -ui -node=server-1 -bootstrap-expect=1 -client="0.0.0.0"