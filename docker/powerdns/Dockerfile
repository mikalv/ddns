FROM pcextreme/powerdns

COPY pdns.conf /etc/powerdns/pdns.conf

COPY entrypoint.sh /
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]

CMD ["pdns_server", "--daemon=no"]