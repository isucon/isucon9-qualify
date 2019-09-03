FROM perl:5.20
RUN mkdir -p /opt/initial-data/result
RUN cpanm -n Crypt::Eksblowfish::Bcrypt Crypt::OpenSSL::Random Digest::SHA JSON JSON::Types;
COPY generator.pl /opt/initial-data
COPY keywords.tsv /opt/initial-data
COPY image_files.txt /opt/initial-data
COPY users.tsv /opt/initial-data
WORKDIR /opt/initial-data
CMD [ "perl", "./generator.pl" ]
