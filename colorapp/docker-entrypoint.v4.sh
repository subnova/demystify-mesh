#!/usr/bin/env sh

set -e

# Redirect traffic to envoy
iptables -t nat -N ENVOY_REDIRECT
iptables -t nat -A ENVOY_REDIRECT -p tcp -j REDIRECT --to-port 15001

# Process inbound traffic
iptables -t nat -A PREROUTING -p tcp -j ENVOY_REDIRECT

# Process outbound traffic
iptables -t nat -N ENVOY_OUTPUT
iptables -t nat -A OUTPUT -j ENVOY_OUTPUT
iptables -t nat -A ENVOY_OUTPUT -p tcp --dport 53 -j RETURN # don't redirect DNS traffic
iptables -t nat -A ENVOY_OUTPUT -m owner --uid-owner 1337 -j RETURN # don't redirect traffic from envoy user
iptables -t nat -A ENVOY_OUTPUT -m owner --gid-owner 1337 -j RETURN # don't redirect traffic from envoy group
iptables -t nat -A ENVOY_OUTPUT -j ENVOY_REDIRECT

cat<<EOF >/etc/resolv.conf
nameserver=127.0.0.1
options ndots:0
EOF

cat<<EOF >/etc/dnsmasq.d/srv.conf
address=/srv/127.0.0.1
server=127.0.0.11
EOF

dnsmasq

su-exec envoy:envoy envoy -c /etc/envoy/envoy.yaml --service-node $(hostname) &
exec su-exec app:app ./colorapp "$@"