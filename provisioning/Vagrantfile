
$disable_unattended_upgrades =<<-SCRIPT
echo "-- Disabling unattended upgrades --"
cat << EOF > /etc/apt/apt.conf.d/51disable-unattended-upgrades
APT::Periodic::Update-Package-Lists "0";
APT::Periodic::Unattended-Upgrade "0";
EOF
SCRIPT

$setup_quarifier_environment =<<-SCRIPT
set -e
echo "-- ISUCON9 quarifier setup --"
export DEBIAN_FRONTEND="noninteractive"
apt update
apt -y install python-pip
pip install ansible==2.8.3
SCRIPT


Vagrant.configure(2) do |config|
  config.vm.box = "bento/ubuntu-18.04"
  config.vm.provider "virtualbox" do |vm|
    vm.name   = "isucon9-quarifier"
    vm.cpus   = 4
    vm.memory = 2048
  end
  config.vm.hostname = "isucon9-quarifier"

  config.vm.provision :shell, inline: $disable_unattended_upgrades
  config.vm.provision :shell, inline: $setup_quarifier_environment
  config.vm.synced_folder "../", "/vagrant", type: "rsync", rsync__exclude: ".git/"
end
