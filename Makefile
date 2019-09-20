.PHONY: prometheus alertmanager node_exporter

prometheus:
	./prometheus/prometheus \
		--config.file="./prometheus/prometheus.yml" \
		--storage.tsdb.path="./prometheus/data" \
		--rules.alert.resend-delay=1s

alertmanager:
	./alertmanager/alertmanager \
		--config.file="./alertmanager/alertmanager.yml" \
		--storage.path="./alertmanager/data"

node_exporter:
	./node_exporter/node_exporter \
		--no-collector.boottime \
		--no-collector.cpu \
		--no-collector.diskstats \
		--no-collector.filesystem \
		--no-collector.loadavg \
		--no-collector.meminfo \
		--no-collector.netdev \
		--no-collector.time \
		--collector.textfile.directory="./node_exporter/desired"

clean:
	rm -rf ./prometheus/data
	rm -rf ./alertmanager/data
