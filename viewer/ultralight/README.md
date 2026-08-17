# ultralight

Ponte cgo para o Ultralight, o motor de HTML/CSS que desenha o HUD.

O SDK nao e versionado (54 MB de DLL). Para reconstruir:

    curl -L -o win.7z   https://ultralight-sdk.sfo2.cdn.digitaloceanspaces.com/ultralight-sdk-latest-win-x64.7z
    curl -L -o linux.7z https://ultralight-sdk.sfo2.cdn.digitaloceanspaces.com/ultralight-sdk-latest-linux-x64.7z

e distribuir assim:

    sdk/include/        cabecalhos (iguais nas duas plataformas)
    sdk/resources/      icudt67l.dat e cacert.pem
    sdk/windows/lib/    *.lib
    sdk/windows/bin/    *.dll
    sdk/linux/bin/      *.so

O binario final precisa de `resources/` ao lado dele. Sem isso o Ultralight
ABORTA o processo sem mensagem em stdout — so o logger conta o que houve.

## O que este WebKit nao suporta

Medido, nao suposto — cada um destes falhou silenciosamente, sem erro e sem log:

| recurso | resultado |
| --- | --- |
| `filter: blur()` | ignorado, a borda fica nitida |
| `backdrop-filter` | ignorado, com e sem `-webkit-` |
| `gap` no flex | ignorado; use `margin` ou `+ irmao` |
| `inset: 0` | ignorado; use `top/right/bottom/left` |
| `image-rendering` | ignorado nas tres palavras-chave |
| data uri acima de ~128 KB | a imagem nunca aparece |
| roda do mouse no documento | so rola elemento com `overflow` |

Funcionam: flex, grid, `var()`, `calc()`, gradientes lineares e radiais,
`transform`, `opacity`, `box-shadow` (inclusive `inset`), `border-radius`,
transicoes, `::-webkit-scrollbar`.

Como nao ha filtro nenhum, vidro aqui e camada translucida com realce interno.
Como `image-rendering` nao existe, o sprite e ampliado em Go antes de virar data
uri, para o engine so reduzir.

## Duas regras que o engine impoe e que nao aparecem em nenhum erro

- **um renderer por processo.** O WebKit guarda estado global que sobrevive a um
  renderer destruido; criar um segundo depois que qualquer view existiu da
  segfault. `Open` devolve sempre o mesmo e nao ha como fechar.
- **o Windows precisa do guard.** As DLLs sao MSVC, e o MSVC batiza uma thread
  disparando a excecao `0x406D1388`, que so um debugger deveria pegar. Sem
  debugger ela chega no handler de ultimo recurso do runtime do Go, que imprime
  um traceback e mata o processo assim que o renderer cria suas threads. O
  `bridge.c` registra um *vectored continue handler* que responde por esse codigo
  e por mais nenhum. Medido: o handler de primeira chance e chamado e a resposta
  dele e ignorada — quem segura e o de continuacao. Nenhum teste cobre isso,
  porque a excecao nao existe fora do Windows; `graft -check` e a verificacao.
- **uma thread por renderer.** Todas as chamadas pertencem a thread que abriu o
  engine. No viewer isso sai de graca, porque a janela ja prende a main; fora
  dela, quem chama tem que se prender sozinho — sem isso quebra em cerca de uma
  execucao a cada sete.
