# Sistema de Monitoramento de Servidor

Aqui vai uma explicação bem simples e direta sobre esse sistema que criei para monitorar o status dos servidores da [FireHosting](https://firehosting.com.br/). É uma ferramenta feita em Go que coleta dados de uso de CPU, RAM, disco e rede a cada 3 minutos, salva tudo no MySQL e ainda estima se o servidor está com carga baixa, média, alta ou crítica.

## O que é esse sistema?

Basicamente, é um monitor que roda no seu servidor e fica de olho nas métricas principais: quanto de CPU, RAM e disco está sendo usado, quantos dados entram e saem pela rede. A cada 3 minutos, ele pega esses dados, calcula uma "carga" do servidor (tipo low, medium, high ou critical) e salva tudo no banco de dados MySQL. Se der algum problema na conexão com o banco, ele guarda os dados localmente num cache e tenta enviar na próxima vez. Super prático para acompanhar o desempenho do seu servidor sem precisar ficar checando manualmente.

## Como funciona?

- **Coleta de dados**: Usa bibliotecas do Go para pegar as métricas do sistema (CPU, RAM, etc.).
- **Estimativa de carga**: Baseado principalmente em CPU e RAM, mas também olha disco e rede. Por exemplo, se CPU ou RAM passar de 80%, marca como "critical".
- **Horário**: Sempre salva com o horário de São Paulo (America/Sao_Paulo).
- **Persistência**: Se o MySQL cair, os dados ficam num arquivo cache.json e são enviados quando voltar.
- **Loop infinito**: Roda para sempre, sem parar, coletando a cada 3 minutos.

## Pré-requisitos

Antes de instalar, certifique-se de que sua máquina tem:
- Linux Ubuntu (ou similar, mas não se limita a ele, podendo funcionar em outros sistemas linux rodando Go).
- Go instalado (versão 1.21 ou superior).
- MySQL rodando, com o banco e tabelas criados (use o db.sql fornecido para cria-lo, pode ser banco local ou remoto).
- Acesso root ou sudo.

## Instalação

Vamos passo a passo para deixar tudo funcionando no Ubuntu.

1. **Clone ou baixe o projeto**: Coloque os arquivos na sua máquina, tipo em `/home/usuario/server_monitor/`.

2. **Instale o Go**: Se não tiver, rode:
   ```
   sudo apt update
   sudo apt install golang-go
   ```
   Verifique com `go version` que deve mostrar algo como `go version go1.21 linux/amd64`.

3. **Instale o MySQL**: Se não tiver, instale e configure:
   ```
   sudo apt install mysql-server
   sudo mysql_secure_installation
   ```
   Crie o banco e as tabelas rodando o `db.sql`:
   ```
   mysql -u root -p < db.sql
   ```

4. **Entre na pasta do projeto**: Vá para a pasta `monitor` dentro do projeto:
   ```
   cd /caminho/para/server_monitor/monitor
   ```

5. **Baixe as dependências**: Rode `go mod tidy` para baixar as bibliotecas necessárias.

6. **Configure o sistema**: Renomeie o arquivo `config.go.example` para `config.go` (se não existir, crie baseado no atual). Abra ele e edite com seus dados:
   - `ServerID`: O ID do seu servidor na tabela `dedicated_servers` (tipo 1 se for o primeiro).
   - `Host`: Endereço do MySQL (normalmente "localhost").
   - `Port`: Porta do MySQL (padrão "3306").
   - `DBName`: Nome do banco (o que você criou com db.sql).
   - `Username`: Usuário do MySQL.
   - `Password`: Senha do usuário.

   Exemplo:
   ```go
   func getConfig() Config {
       return Config{
           ServerID: 1,
           MySQL: MySQLConfig{
               Host:     "localhost",
               Port:     "3306",
               DBName:   "meu_banco",
               Username: "meu_usuario",
               Password: "minha_senha",
           },
       }
   }
   ```

7. **Teste se compila**: Rode `go run .` e veja se aparece "Monitor Running". Se der erro, verifique a config e o banco.

## Como rodar

Para testar, simplesmente rode:
```
go run .
```
Ele vai imprimir "Monitor Running" e começar a coletar dados a cada 3 minutos. Para parar, pressione Ctrl+C.

## Como deixar rodando 24h sem parar

Para deixar o monitor rodando o dia todo sem você precisar ficar logado, use uma dessas opções:

### Opção 1: Usando screen (simples)
1. Instale o screen: `sudo apt install screen`.
2. Rode: `screen -S monitor`.
3. Dentro do screen, vá para a pasta e rode `go run .`.
4. Desconecte com Ctrl+A+D (não fecha o programa).
5. Para voltar: `screen -r monitor`.
6. Para parar: dentro do screen, Ctrl+C.

### Opção 2: Usando systemd (mais profissional e recomendado)
1. Crie um arquivo de serviço: `sudo nano /etc/systemd/system/monitor.service`.
2. Cole isso (ajuste os caminhos):
   ```
   [Unit]
   Description=Server Monitor
   After=network.target

   [Service]
   Type=simple
   User=seu_usuario
   WorkingDirectory=/caminho/para/server_monitor/monitor
   ExecStart=/usr/bin/go run .
   Restart=always
   RestartSec=10

   [Install]
   WantedBy=multi-user.target
   ```
3. Recarregue o systemd: `sudo systemctl daemon-reload`.
4. Inicie: `sudo systemctl start monitor`.
5. Habilite para iniciar automático: `sudo systemctl enable monitor`.
6. Verifique: `sudo systemctl status monitor`.
7. Logs: `sudo journalctl -u monitor -f`.

Com isso, ele roda sozinho e reinicia se cair.

## Interface Web

Além do monitor, criei uma página web simples para visualizar o status dos servidores.

- **API**: Na pasta `web/api/`, tem um `index.php` que retorna dados em JSON sobre os servidores (status, carga, CPU, RAM, uptime).
- **Página de Status**: Na pasta `web/status-page/`, um `index.html` bonito com fundo animado em grid, tema escuro/vermelho (com opção para claro), mostra o status em tempo real, com toggle para ver CPU/RAM.

Para usar:
1. Instale um servidor web como Apache ou Nginx no Ubuntu: `sudo apt install apache2 php libapache2-mod-php`.
2. Copie a pasta `web` para `/var/www/html/` ou configure o servidor para apontar para ela.
3. Acesse `http://localhost/status-page/` para ver a página, e `http://localhost/api/` para a API JSON.
4. Edite o `index.php` com suas credenciais do MySQL (mas mantenha seguro, não exponha!).

A página atualiza automaticamente a cada minuto.
