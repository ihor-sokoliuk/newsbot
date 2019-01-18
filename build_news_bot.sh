#!/bin/bash
echo

program_name="newsbot"
telegram_bot_token=
no_need_in_backup=false

while getopts p:t:b: option
do
case "${option}"
in
p) program_name=${OPTARG};;
t) telegram_bot_token=${OPTARG};;
b) no_need_in_backup=${OPTARG};;
esac
done

program_workdir="/opt/telegram_bot/$program_name"
program_workdir_backup="$program_workdir.backup"
service_name=${program_name}.service
service_file="/etc/systemd/system/$service_name"
backup=false
echo Program name = ${program_name}
echo Program workdir = ${program_workdir}
echo Service file = ${service_file}

echo "---=== Stage 0 ===---"

echo "Backup project..."
if [[ -d ${program_workdir} ]]
then
    if [[ "$no_need_in_backup" == false ]]
    then
        cp -R ${program_workdir} ${program_workdir_backup}
        echo "Backup completed."
        backup=true
    fi
    
	rm -r ${program_workdir}
else
    echo "No need in backup."
fi

echo "---=== Stage 1 ===---"

if [[ -e ${service_file} ]]
then
	echo "Removing service..."
	systemctl stop ${program_name}
  	systemctl disable ${program_name}
  	rm -f ${service_file}
  	systemctl daemon-reload
    echo "Service removed."
else
    echo "Skip removing service."
fi

echo "---=== Stage 2 ===---"

echo "Building project..."
go build -o ${program_name}
mkdir -p ${program_workdir}
cp ${program_name} ${program_workdir}
sed -i 's/{TOKEN}/'${telegram_bot_token}'/g' ${program_name}.yml
cp ${program_name}.yml ${program_workdir}
if [[ "$backup" == true ]]
then
    echo "Restoring backup database and logs..."
	cp ${program_workdir_backup}/${program_name}.log ${program_workdir}
	cp ${program_workdir_backup}/${program_name}.db ${program_workdir}
	echo "Removing backup..."
    rm -r -f ${program_workdir_backup}
fi
echo "Build complited."

echo "---=== Stage 3 ===---"

echo "Creating service..."

/bin/bash -c "echo '[Unit]
Description='$service_name'
Wants=network-online.target
After=network-online.target

[Service]
WorkingDirectory=$program_workdir
ExecStart='$program_workdir/$program_name'
Restart=always
KillMode=process

[Install]
WantedBy=multi-user.target' >> $service_file"

systemctl daemon-reload
systemctl enable ${service_name}
systemctl start ${service_name}

echo "Service creates"
