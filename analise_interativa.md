# Análise Técnica de Mensagens Interativas (Carousel vs Buttons/List)

Realizei uma comparação campo a campo entre a implementação de `SendCarousel` (funcional) e as implementações de `SendButtons` e `SendList` (com falha de entrega). Aqui estão as descobertas fundamentais:

## 1. Divergência na Estrutura de Topo (Oneof)
A principal diferença é o subtipo da mensagem dentro do objeto `InteractiveMessage`:
- **SendCarousel**: Utiliza `CarouselMessage`.
- **SendButtons/SendList**: Utilizam `NativeFlowMessage`.

**Por que isso importa?**
O WhatsApp trata `CarouselMessage` como um contêiner. Mesmo que os cartões dentro dele utilizem `NativeFlowMessage`, o servidor do WhatsApp aceita o contêiner de forma mais permissiva. Mensagens `NativeFlow` puras (fora de um carrossel) são mais rigorosas quanto à versão e ao invólucro.

## 2. O Caso dos Cartões do Carrossel (A Pista Chave)
Os cartões individuais dentro do Carrossel são, na verdade, `NativeFlowMessage`. Se o carrossel chega, sabemos que o seu telefone/cliente suporta `NativeFlow`.

**Diferença sutil no código:**
Nos cartões do carrossel, a `NativeFlowMessage` é montada **sem** os campos `MessageVersion` e **sem** o objeto `ContextInfo`.

## 3. A "Armadilha" do Rodapé (Footer)
Notei uma diferença na forma como o rodapé é tratado:
- No `SendCarousel`, o objeto `Footer` só é adicionado se o texto não for vazio.
- No `SendButtons` e `SendList`, o objeto `Footer` está sendo enviado sempre, mesmo que o texto seja uma string vazia (`""`).

**Impacto**: Alguns clientes do WhatsApp rejeitam mensagens interativas se um objeto opcional (como Footer) for enviado vazio.

## 4. Comparativo de Envelopamento (Wrapper)
- No `SendCarousel`, a mensagem é enviada como um `waE2E.Message` simples.
- Em minhas tentativas anteriores (v4), tentei envolver em `ViewOnceMessage`. É possível que a versão do seu cliente de WhatsApp exija o formato direto (como o Carousel) mas com configurações de versão específicas.

## 5. JSON dos Parâmetros (ButtonParamsJSON)
- O `SendCarousel` usa `fmt.Sprintf` para montar o JSON manualmente.
- O `SendButtons` usa `json.Marshal`. Embora `Marshal` seja tecnicamente mais correto, o WhatsApp às vezes é sensível a escapes de caracteres que o `Marshal` faz e o `Sprintf` não.

---

# Proposta de Ajuste (Refinamento v6)

Para alinhar 100% com o padrão de sucesso do Carousel, devemos aplicar estas correções cirúrgicas:

1. **Condicional de Footer**: Remover o objeto `Footer` completamente se o texto estiver vazio.
2. **Remoção de ContextInfo**: O Carousel não usa, então os outros também não devem usar (já ajustado na v5, mas manteremos).
3. **Alinhamento de JSON**: Usar a mesma técnica de montagem de JSON do Carousel para garantir que caracteres especiais não causem rejeição.
4. **Simplificação do Objeto de Lista**: Garantir que as linhas da lista contenham apenas os campos `id`, `title` e `description` (sem campos extras de compatibilidade que adicionei antes).
