# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure(2) do |config|
  config.vm.box = 'bento/ubuntu-14.04'
  config.vm.network 'forwarded_port', guest: 2376, host: 2376 // docker
  config.vm.network 'forwarded_port', guest: 2379, host: 2379 // zfs-freight
  config.vm.network 'forwarded_port', guest: 4500, host: 4500 // consul

  config.vm.provider 'virtualbox' do |vb|
    vb.memory = '1024'
  end

  config.omnibus.chef_version = '12.6.0'
  config.berkshelf.enabled = true

  config.vm.provision :chef_solo do |chef|
    chef.cookbooks_path = 'cookbooks'

    chef.json = {
      go: {
        gopath: '/vagrant'
      }
    }

    chef.run_list = [
      'recipe[golang::default]',
      'recipe[zfs-freight::default]'
    ]
  end
end
