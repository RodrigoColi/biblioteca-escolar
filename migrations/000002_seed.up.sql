INSERT INTO alunos (matricula, nome, email, turma) VALUES
    (1001, 'Ana Souza',       'ana.souza@escola.com',       '9A'),
    (1002, 'Bruno Lima',      'bruno.lima@escola.com',      '9A'),
    (1003, 'Carla Mendes',    'carla.mendes@escola.com',    '8B'),
    (1004, 'Diego Ferreira',  'diego.ferreira@escola.com',  '8B'),
    (1005, 'Eduarda Rocha',   'eduarda.rocha@escola.com',   '7C');

INSERT INTO livros (titulo, autor, genero, ano_publicacao) VALUES
    ('Dom Casmurro',             'Machado de Assis',     'Romance',  1899),
    ('O Cortiço',                'Aluísio Azevedo',      'Romance',  1890),
    ('O Pequeno Príncipe',       'Antoine de Saint-Exupéry', 'Infantojuvenil', 1943),
    ('A Revolução dos Bichos',   'George Orwell',        'Ficção',   1945),
    ('Vinte Mil Léguas Submarinas', 'Júlio Verne',       'Aventura', 1870),
    ('Sítio do Picapau Amarelo', 'Monteiro Lobato',      'Infantojuvenil', 1920);

INSERT INTO emprestimos (matricula, id_livro, data_emprestimo, data_prevista, data_devolucao) VALUES
    (1001, 1, '2026-09-01', '2026-09-15', '2026-09-14'),
    (1002, 3, '2026-09-10', '2026-09-24', '2026-09-30'),
    (1003, 4, '2026-10-01', '2026-10-15', NULL),
    (1004, 5, '2026-10-20', '2026-11-03', NULL),
    (1001, 6, '2026-10-22', '2026-11-05', NULL),
    (1005, 2, '2026-09-20', '2026-10-04', NULL);
