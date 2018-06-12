FROM cassandra

RUN sed -i -r 's/authenticator: AllowAllAuthenticator/authenticator: PasswordAuthenticator/' /etc/cassandra/cassandra.yaml
RUN sed -i -r 's/authorizer: AllowAllAuthorizer/authorizer: CassandraAuthorizer/' /etc/cassandra/cassandra.yaml
RUN sed -i -r 's/enable_user_defined_functions: false/enable_user_defined_functions: true/' /etc/cassandra/cassandra.yaml
RUN sed -i -r 's/enable_scripted_user_defined_functions: false/enable_scripted_user_defined_functions: true/' /etc/cassandra/cassandra.yaml
