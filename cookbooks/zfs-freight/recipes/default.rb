#
# Cookbook Name:: zfs-freight
# Recipe:: default
#
# Copyright (c) 2016 The Authors, All Rights Reserved.

include_recipe 'zfs_linux::default'

bash 'create-sparse-file' do
  user  'root'
  group 'root'

  code <<-EOF
    dd if=/dev/zero of=/zfs.img bs=1 count=0 seek=15G
  EOF

  not_if { ::File.exists?('/zfs.img') }
  action :run
end

bash 'create-zpool' do
  user  'root'
  group 'root'

  code <<-EOF
    zpool create tank /zfs.img
  EOF

  not_if 'zpool list | grep -q tank'
  action :run
end

bash 'create-docker-zfs' do
  user  'root'
  group 'root'

  code <<-EOF
    zfs create -o mountpoint=/var/lib/docker -o compression=on -o atime=off tank/docker
  EOF

  not_if 'zfs list | grep -q tank/docker'
  action :run
end

include_recipe 'chef-apt-docker::default'

docker_service 'default' do
  install_method 'package'
  version        '1.11.0'

  host [
    'unix:///var/run/docker.sock',
    'tcp://0.0.0.0:2376'
  ]

  iptables true

  storage_driver 'zfs'

  log_driver 'json-file'
  log_opts   [
    'max-file=10',
    'max-size=10m'
  ]

  action [:create, :start]
end

docker_volume 'consul' do
  action :create
end

docker_image 'consul' do
  repo 'consul'
  tag  'v0.6.4'

  read_timeout  600
  write_timeout 600

  action :pull_if_missing
end

docker_container 'consul' do
  repo 'consul'
  tag  'v0.6.4'

  port [
    '0.0.0.0:8300:8300/tcp',
    '0.0.0.0:8301:8301/tcp',
    '0.0.0.0:8301:8301/udp',
    '0.0.0.0:8400:8400/tcp',
    '0.0.0.0:8500:8500/tcp',
    '0.0.0.0:8600:8600/tcp',
    '0.0.0.0:8600:8600/udp'
  ]

  env [
    'CONSUL_LOCAL_CONFIG={"leave_on_terminate": true}'
  ]

  volumes [
    'consul:/consul/data:rw'
  ]

  command <<-EOF
    consul agent -server -ui \
    -bootstrap \
    -data-dir=/consul/data \
    -client=0.0.0.0 \
    -advertise=#{node['ipaddress']} \
    -node=#{node['fqdn']} \
    -recursor=8.8.8.8
  EOF

  restart_policy 'always'

  action :run
end

node.override['dnsmasq']['dns']['server'] = '127.0.0.1#8600'

include_recipe 'dnsmasq::default'
include_recipe 'dnsmasq::dns'
