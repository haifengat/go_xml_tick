FROM haifengat/ctp_real_md

COPY bin/xml_tick /home
RUN chmod a+x /home/xml_tick
ENTRYPOINT ["/home/xml_tick"]
