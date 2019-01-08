#!/bin/bash
program_name="NewsBot"
program_workdir="/home/user/programs/$program_name"
program_workdir_backup="/home/user/programs/$program_name.backup"
service_file="/etc/systemd/system/$program_name.service"
backup=false
echo Program name = $program_name
echo Program workdir = $program_workdir
echo Service file = $service_file

echo "---=== Stage 0 ===---"

echo Backup project...
if [[ -d ${program_workdir} ]]
then
	cp -R $program_workdir $program_workdir_backup
	rm -r $program_workdir
    backup=true
fi
if [[ "$backup" == true ]]
then
	echo Backup completed.
else
	echo No need in backup.
fi

echo "---=== Stage 1 ===---"


echo Clean project...
if [[ -e ${service_file} ]]
then
	echo Removing service
	systemctl stop ${program_name}
  	systemctl disable ${program_name}
  	rm -f ${service_file}
  	systemctl daemon-reload
fi
if [[ -d ${program_workdir} ]]
then
	echo Removing workdir
    rm -r -f ${program_workdir}
fi
echo Clean completed.

echo "---=== Stage 2 ===---"

echo Build project...
sed -i 's/{TOKEN}/'$telegram_bot_token'/g' ${program_name}.yml
echo Dep ensure...
echo GOPATH=$GOPATH
echo PATH=$PATH
go get -u github.com/golang/dep/cmd/dep
dep ensure
echo Building...
go build -o ${program_name}
mkdir -p ${program_workdir}
cp ${program_name} ${program_workdir}
cp ${program_name}.yml ${program_workdir}
if [[ "$backup" == true ]]
then
	cp ${program_workdir_backup}/${program_name}.log ${program_workdir}
	cp ${program_workdir_backup}/${program_name}.db ${program_workdir}
fi
echo Build complited.

echo "---=== Stage 3 ===---"

echo Create service...

/bin/bash -c "echo '[Unit]
Description='$program_name'
Wants=network-online.target
After=network-online.target

[Service]
ExecStart='$program_workdir/$program_name'
Restart=always
KillMode=process

[Install]
WantedBy=multi-user.target' >> $service_file"

systemctl daemon-reload
systemctl enable ${program_name}
systemctl start ${program_name}

echo Create completed.
