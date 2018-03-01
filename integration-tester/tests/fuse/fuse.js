const kelda = require('kelda');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();

// The SSH key used to authenticate the SSH connection. Although it's ugly to
// save them here in plaintext, it's safe to do so as all traffic is in the
// private network, and the SSH server isn't hosting anything of importance.
const sshPubKey = 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCsxTbw33l7enItiAptl' +
  'OkIj+N6Mg1HZxXXrX/V/PcODNEeVvDxokhfM1vJnPhiaYu+qLIvDJeixo9SO0xIbKcJ9p74af' +
  'GsfEShfw4MPCSaaJk/gni/tqrNZxBf7AgTkO/96uPToALeXkAFVw7AtwlV16RqCB7+kUMiW62' +
  'o9hmLCbjUuI7myJCMjCffQaLC7Oz7cCphxEG1lNb4E7dz/6CEJQF+84EuRfLjuq6sJchIHs2j' +
  'NriV/vL6PXB9a8DAzTCvzrHlf3A4gESn2CkEScBq+QaC3EN4sM2C8LqlQAfFFSi5N5siHmDBb' +
  'zHYlPDLnLj2iL0IXveTno5RjbjZbYd1';
const sshPrivKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEArMU28N95e3pyLYgKbZTpCI/jejINR2cV161/1fz3DgzRHlbw
8aJIXzNbyZz4YmmLvqiyLwyXosaPUjtMSGynCfae+GnxrHxEoX8ODDwkmmiZP4J4
v7aqzWcQX+wIE5Dv/erj06AC3l5ABVcOwLcJVdekagge/pFDIlutqPYZiwm41LiO
5siQjIwn30Giwuzs+3AqYcRBtZTW+BO3c/+ghCUBfvOBLkXy47qurCXISB7Noza4
lf7y+j1wfWvAwM0wr86x5X9wOIBEp9gpBEnAavkGgtxDeLDNgvC6pUAHxRUouTeb
Ih5gwW8x2JTwy5y49oi9CF73k56OUY242W2HdQIDAQABAoIBAQCg34IpB+22bG2k
t+f94Yqbzl+iiLiUpAhSq9s9Bi+Fhamy9oGkfdelzczKDr+5402cWriP1snbZ8hM
aaq+RW3EeT/NT9kZIx1Ew2nxOo9at8r6uCJ3YT/pwu4cY0uh7HOLnAxIIdaJ+Hjk
gAvcppKqvAD8OlOh9uDsPTGXApEGRJ4w96UFbnUOqv4IwuTy1nF01w71JrBafs8T
v6gxsyQpiFewz3Btb0lAkidQfL1c8Q6YgdpPLi1lC5I9l7K6WFxsQGgZcO7vXyFn
m0Ppwij3y2llLyK84+XDNhFdvWuw2aKncLeHY78NsyqOgbpFLHeZYMnR9EWuRpUn
3T4yDNCBAoGBAOSdEtNVx44yAESqZ3qHmOLnMivllUnZOC6VTRsRVPs03aon4Tf3
/f24kbmgaHE8/Z8CZ4+uC6q9JJHHRovaBFlvkciTB0MZfQ0PtkIDEtggxZo2/khe
NgyFDxJWrFKJ7/ltW3ETdAeGRTZW294i2Lku8o8gedYxP61w6ElCfJBfAoGBAMF3
lt+wmniiGOxAhwDxHJRNIArIvrf03VseYnuxFWku37N1g7craJ4c8xfNs+KMNc6a
1WPmtN+3sJmsK3KTxrtPJvHOIsNWQulv36wvgKl9ZVstj0gbqOwNvVNh+uaSg75k
4c6JhFS6WWcHG7RCCsIjTlimhu1r/A/NPPhRfeirAoGADP9ZKKbB17XECiNeCrtW
19+pHJHK8Q3mgc9/OMC9giK5T4lA5ru0tw4dSt5x0a5UBQxP8v1EMIrcX2Vi/2R/
xs3vDeY+DXSPhYSVKh+enKcQVPo3bsncbM3L05EV7wNkn1u2TTF78UmS+cnqajC0
/aJLrBN+mczm/+dhbXjYOCUCgYEAsVKPMo+HxbGs7j1mf/J+o17dU4UTaUBB8tYy
pfR1D2crGi1HgIeE6AbYuKSNj8O7PZakp2A5wCN49iDb4bSYne26YD7zld5mjddA
R21ym+aXE676eLkBZvpg4SAY+2Sm48dLQCbC53W1o7zcI6e0fKQnlxFq8gnbihAv
JdprcOkCgYBebUPExYATSgiPDNCPfZHdjraWA6UdLxtXMwpDIFz4kQgz1l4C3g7Y
DR2Ly5emUiIc0GT6yoeCWhmhjoJJ3xNNmooG6+b5F574tAJKdP7Kz4xwbZwyx61w
upFf/aftG79xRDxTYovMJEsZp24cfT0RuY0nkFh1GYaFTh3n7yb+pQ==
-----END RSA PRIVATE KEY-----`;

// Set up the SSH server.
const sshServerImage = new kelda.Image({
  name: 'ssh-server',
  dockerfile: `FROM ubuntu:16.04
RUN apt-get update && apt-get install -y openssh-server
RUN sed -ri 's/^PermitRootLogin\\s+.*/PermitRootLogin yes/' /etc/ssh/sshd_config
RUN mkdir -p /var/run/sshd`,
});

const sshServer = new kelda.Container({
  name: 'ssh-server',
  image: sshServerImage,
  command: ['/usr/sbin/sshd', '-D'],
  filepathToContent: {
    '/root/.ssh/authorized_keys': sshPubKey,
  },
});

// Set up the containers that will attempt to mount from SSH server.
const sshfsImage = new kelda.Image({
  name: 'sshfs',
  dockerfile: `FROM ubuntu:16.04
RUN apt-get update && apt-get install -y sshfs`,
});

const sshfsContainerFields = {
  image: sshfsImage,
  command: ['sh', '-c',
    'chmod 0400 ~/.ssh/id_rsa && ' +
    'tail -f /dev/null'],
  filepathToContent: {
    '/root/.ssh/id_rsa': sshPrivKey,
  },
};

const canMount = new kelda.Container(Object.assign({}, sshfsContainerFields, {
  name: 'can-mount',
  privileged: true,
}));

const cannotMount = new kelda.Container(Object.assign({}, sshfsContainerFields, {
  name: 'cannot-mount',
}));

// Allow the sshfs containers to connect to the SSH server.
kelda.allowTraffic([canMount, cannotMount], sshServer, 22);

// Deploy the containers.
canMount.deploy(infra);
cannotMount.deploy(infra);
sshServer.deploy(infra);
