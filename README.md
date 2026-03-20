# ⏱ hourly-cli

CLI para registro e controle de horas trabalhadas. Calcule automaticamente o tempo por dia, semana ou mês — com suporte a valor/hora, notas e importação de CSV do Jira.

---

## Sumário

- [⏱ hourly-cli](#-hourly-cli)
  - [Sumário](#sumário)
  - [Instalação](#instalação)
    - [Pré-requisitos](#pré-requisitos)
    - [Build a partir do fonte](#build-a-partir-do-fonte)
  - [Início rápido](#início-rápido)
  - [Comandos](#comandos)
    - [`add`](#add)
    - [`list`](#list)
    - [`report`](#report)
    - [`delete`](#delete)
    - [`config`](#config)
      - [`config set`](#config-set)
      - [`config show`](#config-show)
    - [`import`](#import)
  - [Armazenamento](#armazenamento)
  - [Formato de data e hora](#formato-de-data-e-hora)
  - [Licença](#licença)

---

## Instalação

### Pré-requisitos

- [Go 1.22+](https://go.dev/dl/)

### Build a partir do fonte

```bash
git clone https://github.com/suhai-art/hourly-cli.git
cd hourly-cli
./build.sh
```

O binário será gerado em `./build/hourly`.

Opcionalmente, mova-o para algum diretório do seu `$PATH`:

```bash
mv ./build/hourly /usr/local/bin/hourly
```

---

## Início rápido

```bash
# Registra entrada agora
hourly add 09:00

# Registra entrada e saída
hourly add 09:00 18:00

# Lista as horas de hoje
hourly list

# Vê o relatório do mês
hourly report

# Configura valor por hora
hourly config set 75.50 --currency "R$"
```

---

## Comandos

### `add`

Registra um período de trabalho. Aceita apenas entrada (em aberto) ou entrada + saída.

```
hourly add <entrada> [saída] [--note "texto"]
```

| Argumento | Descrição |
|-----------|-----------|
| `entrada` | Horário de início (`HH:MM` ou `YYYY-MM-DD HH:MM`) |
| `saída` | Horário de fim (opcional) |

| Flag | Atalho | Descrição |
|------|--------|-----------|
| `--note` | `-n` | Nota opcional associada ao registro |

**Exemplos:**

```bash
# Registra apenas a entrada (fica em aberto)
hourly add 09:00

# Registra entrada e saída no dia atual
hourly add 09:00 18:00

# Registra período em data específica com nota
hourly add "2024-03-15 08:30" "2024-03-15 17:45" --note "reunião manhã"
```

---

### `list`

Lista os registros de horas. Por padrão exibe os registros de hoje.

```
hourly list [--week] [--month] [--day YYYY-MM-DD]
```

| Flag | Atalho | Descrição |
|------|--------|-----------|
| `--week` | `-w` | Exibe registros da semana atual |
| `--month` | `-m` | Exibe registros do mês atual |
| `--day` | `-d` | Exibe registros de uma data específica (`YYYY-MM-DD`) |

**Exemplos:**

```bash
hourly list                    # hoje
hourly list --week             # semana atual
hourly list --month            # mês atual
hourly list --day 2024-03-15   # dia específico
```

A listagem exibe data, horários de entrada/saída, duração e — caso configurado — o valor ganho no período.

---

### `report`

Exibe um relatório consolidado de horas por dia no mês, com total geral.

```
hourly report [--month YYYY-MM]
```

| Flag | Atalho | Descrição |
|------|--------|-----------|
| `--month` | `-m` | Mês de referência no formato `YYYY-MM` (padrão: mês atual) |

**Exemplos:**

```bash
hourly report                  # mês atual
hourly report --month 2024-03  # março de 2024
```

---

### `delete`

Remove registros de horas.

```
hourly delete [id] [--all]
```

| Argumento | Descrição |
|-----------|-----------|
| `id` | ID do registro a remover (opcional) |

| Flag | Descrição |
|------|-----------|
| `--all` | Remove **todos** os registros (solicita confirmação) |

**Modos de uso:**

```bash
hourly delete                    # abre seletor interativo (multi-seleção)
hourly delete 20240315090000     # remove pelo ID
hourly delete --all              # remove todos os registros
```

No modo interativo, use `espaço` para marcar registros e `enter` para confirmar.

---

### `config`

Gerencia a configuração de valor/hora e moeda.

#### `config set`

Define o valor cobrado por hora.

```
hourly config set <valor_por_hora> [--currency "símbolo"]
```

| Flag | Atalho | Descrição |
|------|--------|-----------|
| `--currency` | `-c` | Símbolo da moeda (ex: `R$`, `USD`, `€`) |

```bash
hourly config set 50
hourly config set 75.50 --currency "USD"
hourly config set 100 --currency "€"
```

#### `config show`

Exibe a configuração atual.

```bash
hourly config show
```

---

### `import`

Importa registros de horas a partir de um CSV exportado pelo Jira.

```
hourly import <arquivo.csv> [--dry-run]
```

| Flag | Descrição |
|------|-----------|
| `--dry-run` | Exibe o que seria importado sem salvar nada |

**Colunas esperadas no CSV:**

| Coluna | Obrigatória | Descrição |
|--------|-------------|-----------|
| `Date` | ✅ | Data no formato `YYYY-MM-DD` |
| `Time Seconds` | ✅* | Duração em segundos |
| `Time` | ✅* | Alternativa: duração em formato humano (ex: `8h`, `1h 30m`) |
| `Member` | ❌ | Nome do membro |
| `Project` | ❌ | Nome do projeto |
| `Issue` | ❌ | Chave da issue |
| `Comment` | ❌ | Comentário |

> \* `Time Seconds` ou `Time` — pelo menos um deve estar presente.

**Comportamento de merge:** se já existir um registro com `note: "jira-import"` no mesmo dia, as horas são somadas em vez de criar um novo registro.

**Exemplos:**

```bash
hourly import jira.csv
hourly import jira.csv --dry-run
```

---

## Armazenamento

Todos os dados são persistidos localmente em `~/.hourly/`:

| Arquivo | Conteúdo |
|---------|----------|
| `~/.hourly/data.json` | Registros de horas |
| `~/.hourly/config.json` | Configurações (valor/hora, moeda) |

Os arquivos são JSON legíveis — você pode inspecioná-los ou editá-los manualmente se necessário.

---

## Formato de data e hora

O comando `add` aceita dois formatos de horário:

| Formato | Exemplo | Comportamento |
|---------|---------|---------------|
| `HH:MM` | `09:00` | Usa a data de hoje |
| `YYYY-MM-DD HH:MM` | `2024-03-15 08:30` | Data e hora completas |

---

## Licença

MIT — sinta-se livre para usar, modificar e distribuir.