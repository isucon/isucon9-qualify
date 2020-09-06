require 'mysql2'
require 'newrelic_rpm'

class NrMysql2Client < Mysql2::Client
  def initialize(*args)
    super
  end

  def query(sql, *_args)
    if ENV['LOCAL']
      puts sql
      puts caller(0)[1]
    end
    callback = lambda do |_result, metrics, elapsed|
      NewRelic::Agent::Datastores.notice_sql(sql, metrics, elapsed)
    end
    op = sql[/^(select|insert|update|delete|begin|commit|rollback)/i]
    table = sql[/\busers|events|sheets|reservations|administrators\b/] || 'others'
    NewRelic::Agent::Datastores.wrap('MySQL', op, table, callback) do
      super
    end
  end
end
