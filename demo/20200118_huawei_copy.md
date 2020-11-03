# Huawei copy

-----------------------------

## [Huawei Homepage](20200118_huawei.md)

## [Huawei Docker images](20200118_huawei_dockerimage.md)

## [Huawei 文件拷贝](20200118_huawei_copy.md)

## [Huawei GCS 流程设计](20200118_huawei_gcspipeline.md)

-----------------------------

## 1. Obs-util

### [安装界面](https://support.huaweicloud.com/utiltg-obs/obs_11_0003.html)

```sh
cd /Users/xugang/Desktop/sequencing_center/d-huawei
#./obsutil config -i=ak -k=sk -e=endpoint

# Set the environment 
./obsutil config -i=5ULAG** -k=gvroYZE9uUmp3igpEPAEQ*****Qn9kBoHz02 -e=https://obs.cn-north-4.myhuaweicloud.com

#配置完成后，您可以通过如下方式检查连通性，确认配置是否无误
./obsutil ls obs://hw-gcs-logo-cn-north-4-06a54be3938010610f01c00da675d700/

```


### 1. 在北京4创建一个名字为 bucket-test-xg的新桶

```sh
./obsutil mb obs://bucket-test-xg -location=cn-north-4
```

* 华北-北京四	cn-north-4	obs.cn-north-4.myhuaweicloud.com	HTTPS/HTTP

### 2. 将文件test.txt 上传到 bucket-test-xg 桶中。

```sh
./obsutil cp ./test.txt obs://bucket-test-xg/test.txt
```

### 3. 运行./obsutil cp obs://bucket-test/test.txt /temp/test1.txt命令，将bucket-test-xg桶中的test.txt对象下载至本地。

```sh
./obsutil cp obs://bucket-test-xg/test.txt ./test1.txt

./obsutil cp obs://hw-gcs-logo-cn-north-4-06a54be3938010610f01c00da675d700/test1.txt ./test1.txt

```

### 4. 运行./obsutil rm obs://bucket-test/test.txt -f命令，在bucket-test桶中删除test.txt对象。

```sh

./obsutil rm obs://bucket-test-xg/test.txt -f

```

### 5. 运行./obsutil rm obs://bucket-test -f命令，删除bucket-test桶。

```sh
./obsutil rm obs://bucket-test-xg -f

```

### 6. 列举桶

```sh
./obsutil ls -limit=5
```

### 7. 在桶中创建文件夹

```sh
./obsutil mkdir obs://bucket/folder[/subfolder1/subfolder2] [-config=xxx]

./obsutil mkdir obs://gene-container-xugang/test-xg

./obsutil mkdir obs://hw-gcs-logo-cn-north-4-06a54be3938010610f01c00da675d700/arabidopsis/huawei_file/Ribocode

```

### 8.上传对象

上传单个或多个本地文件或文件夹至OBS指定路径。待上传的文件可以是任何类型：文本文件、图片、视频等等。

**约束与限制**
obsutil对上传的文件或文件夹有大小限制，最小可以上传0Byte的空文件或文件夹，最大可以上传5GB（未采用分段上传）或48.8TB（采用分段上传）的单个文件或文件夹。

```sh
#上传文件
./obsutil cp file_url obs://bucket[/key] 

./obsutil cp test.txt obs://gene-container-xugang/test-xg

#上传文件夹
./obsutil cp folder_url obs://bucket[/key] 

./obsutil cp ./temp obs://gene-container-xugang/test-xg -f -r


#多文件/文件夹上传
./obsutil cp file1_url,folder1_url|filelist_url obs://bucket[/prefix] 
```

```sh
for i in `ls|grep fq$`;
do echo $i;
	echo /lulab/lustre2/xugang/docker_backup/huawei/obsutil_linux_amd64_5.1.11/obsutil cp ${i} obs://hw-gcs-logo-cn-north-4-06a54be3938010610f01c00da675d700/arabidopsis/huawei_file/Ribocode
/lulab/lustre2/xugang/docker_backup/huawei/obsutil_linux_amd64_5.1.11/obsutil cp ${i} obs://hw-gcs-logo-cn-north-4-06a54be3938010610f01c00da675d700/arabidopsis/huawei_file/Ribocode
done
```

|参数|约束|描述|
|-|-|-|
|r|附加参数，上传文件夹时必选 多文件上传时可选|递归上传文件夹中的所有文件和子文件夹。|
|f|附加参数，多文件/文件夹上传或上传文件夹时可选|强制操作，不进行询问提示。|

### 9.查询对象属性

```sh
./obsutil stat obs://gene-container-xugang/test-xg
```

### 10.设置对象属性。

```sh

obsutil chattri obs://bucket-test/key -acl=public-read

obsutil chattri obs://bucket-test -r -f -acl=public-read

```

* private
* public-read
* public-read-write
* bucket-owner-full-control

* 说明： 以上四个值分别对应：私有读写、公共读、公共读写、桶拥有者完全控制，四种预定义访问策略。

### 11.列举对象

```sh
./obsutil ls obs://bucket[/prefix] [-s] [-d] [-v] [-marker=xxx] [-versionIdMarker=xxx] [-bf=xxx] [-limit=1] [-config=xxx]

```

## obs example
```sh

obsutil config -i=5ULAGR0********8Y6P -k=gvroYZE9uUmp3i********Hz02 -e=https://obs.cn-north-4.myhuaweicloud.com && obsutil ls && obsutil cp -r -f -u obs://gene-container-xugang/gcs/ /home/sfs && ls /home/sfs 

obsutil config -i=${gcs_id} -k=${gcs_password} -e=${http} && obsutil cp -r -f -u ${obs_data} /home/sfs && obsutil cp -r -f -u ${obs_reference} /home/sfs && ls /home/sfs

obsutil config -i=5ULAGR0********6P -k=gvr*************BoHz02 -e=https://obs.cn-north-4.myhuaweicloud.com && obsutil cp /home/sfs/ obs://gene-container-xugang/gcs/output -r -f && rm -rf /home/sfs && echo Check sfs && ls -alh /home/sfs 

      - 'obsutil config -i=5ULA********V57Y6P -k=g*****************88oHz02 -e=https://obs.cn-north-4.myhuaweicloud.com -e=https://obs.cn-north-4.myhuaweicloud.com && obsutil cp /home/sfs/ obs://gene-container-xugang/gcs/output -r'
      - 'rm -rf /home/sfs/*' 
      - 'echo Check sfs' 
      - 'ls -al /home/sfs'

volumes:
  volumes-4ndk:
    mount_path: '/home/sfs'
    mount_from:
      pvc: '${GCS_SFS_PVC}'

```
```sh
GCS_SFS_PVC
GCS_DATA_PVC
GCS_REF_PVC
```

