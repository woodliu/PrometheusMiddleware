FROM centos:7
ADD prometheusservice /usr/local/bin/prometheusservice
CMD ["prometheusservice"]
