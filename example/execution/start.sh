if [ ! -d $GOPATH/src/github.com/kubernetes-incubator ]
then
    echo "required directory doesn't exist in GOPATH so creating the same"
    mkdir -p $GOPATH/src/github.com/kubernetes-incubator    
else
    echo "required directory exist"	
fi

echo "move to the kubernetes-incubator folder"
cd $GOPATH/src/github.com/kubernetes-incubator

echo "start cloning the external-storage"
git clone https://github.com/kubernetes-incubator/external-storage.git


echo "move to the nfs folder then build by just make"
cd ./external-storage/nfs/

make

sudo ./nfs-provisioner -provisioner=example.com/nfs  -kubeconfig=$HOME/.kube/config  -run-server=false -use-ganesha=false -server-hostname=100.64.41.154
