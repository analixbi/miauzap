# Miauzap para WooCommerce

Plugin WordPress/WooCommerce para integrar a API Miauzap com:

- Login por OTP via WhatsApp para clientes ja cadastrados.
- Fluxo de mensagens por status de pedido, incluindo status customizados registrados no WooCommerce.
- Editor com variaveis de pedido/produto por clique ou arrasta e solta.
- Previsualizacao de mensagens usando pedido real.
- Multiplas instancias Miauzap com rotacao ponderada.
- Fila automatica com atraso aleatorio, tentativas, limite diario por instancia, Action Scheduler/WP-Cron e horario silencioso.
- Worker interno para acordar a fila em segundo plano quando houver mensagens agendadas nos proximos minutos.
- Marketing para novos produtos, volta de estoque e campanhas manuais.
- Resolucao opcional de contato pelo endpoint `/user/check`, usando o `JID` retornado pela API no envio.
- Tratamento de telefone brasileiro com e sem nono digito, respeitando primeiro o formato informado no pedido.
- Variaveis dinamicas de anotacoes do pedido por palavra-chave.
- Variaveis de campos personalizados/meta do pedido, como codigo de rastreio dos Correios.

## Fluxo de mensagens por status

Cada status de pedido pode ter uma ou mais mensagens. A primeira mensagem usa apenas o intervalo randomizado geral da fila. As mensagens seguintes podem ter um campo **Aguardar antes desta mensagem**, somado ao intervalo randomizado entre disparos.

Exemplo de uso:

```text
Mensagem 1: Seu pedido {numero_pedido} foi postado.
Aguardar: 8 segundos
Mensagem 2: {correios_tracking_code}
```

Assim o codigo de rastreio pode chegar em uma mensagem separada, mais facil de copiar no WhatsApp.

## Endpoints Miauzap usados

- `POST /chat/send/text`
  - Headers: `Token` e `Authorization` com o token da instancia.
  - Body: `{"number":"5511999999999","text":"mensagem","linkPreview":true}`
- `POST /user/check`
  - Body: `{"Phone":["5511999999999"]}`
  - O plugin usa o campo `JID` retornado para disparar a mensagem ao contato confirmado pelo WhatsApp.
- `GET /session/status`
  - Usado para teste de instancia.
- `GET /health`
  - Usado para teste basico da API.

## Instalacao

1. Copie a pasta `miauzap-woocommerce` para `wp-content/plugins/`.
2. Ative o plugin no WordPress.
3. Acesse **Miauzap > Conexao**.
4. Informe a URL base da API e uma ou mais instancias com token.
5. Configure atrasos, limites e textos.

## Shortcode

Use o shortcode abaixo em qualquer pagina:

```text
[miauzap_otp_login]
```

O formulario tambem aparece no fim do login padrao do WooCommerce quando OTP estiver ativo. Ele usa o campo **Telefone de cobranca** do cliente, aceitando DDD + numero no formulario.

## Variaveis de anotacoes do pedido

Nas mensagens de status, use anotacoes do pedido:

```text
{nota:palavra}
{nota_cliente:palavra}
{nota_privada:palavra}
{observacao:palavra}
```

O plugin procura a anotacao mais recente que contenha a palavra. Se a anotacao tiver formato `palavra: texto`, ele usa somente o texto depois dos dois pontos.

Exemplo:

```text
Ola {primeiro_nome}, o prazo combinado foi: {nota:prazo}
```

Com uma anotacao `prazo: entrega ate sexta-feira`, a mensagem recebe `entrega ate sexta-feira`.

Para o caso de remessa mostrado no WooCommerce:

```text
Sua remessa JET: {nota:Remessa_JET}
```

Com a anotacao `Remessa_JET:888030755699760`, a mensagem enviada fica com `888030755699760`.

## Variaveis de campos personalizados do pedido

Para usar uma meta/campo personalizado do pedido:

```text
{meta:nome_do_campo}
```

Para o rastreio dos Correios salvo em `_correios_tracking_code`, use qualquer uma destas:

```text
{correios_tracking_code}
{codigo_rastreio}
{meta:_correios_tracking_code}
```

## Observacoes de seguranca operacional

Este plugin foi desenhado para reduzir risco de bloqueio:

- Nao dispara em massa imediatamente quando varios pedidos mudam de status.
- Espaca os envios pela fila com minimo e maximo configuraveis.
- Quando varios itens ficam vencidos ao mesmo tempo, reespaca o backlog antes de continuar.
- Agenda cada item da fila automaticamente e tenta acordar o WP-Cron quando houver mensagem vencida.
- Permite horario silencioso.
- Pode consultar o contato no WhatsApp antes de enviar e usar o JID retornado pela API.
- Permite distribuir envios entre varias instancias.
- Campanhas de marketing respeitam opt-in por padrao.

Mesmo assim, boas praticas continuam importantes: envie somente para clientes com relacionamento/consentimento, personalize textos e evite volumes agressivos em numeros novos.
