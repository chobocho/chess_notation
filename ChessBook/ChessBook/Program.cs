namespace ChoboChessBoard;

class ChessBook
{
    private ChessBoard _chessBoard;

    private string chessGame =
        "1.e4 e5 2.f4 exf4 3.Bc4 Qh4+ 4.Kf1 b5 5.Bxb5 Nf6 6.Nf3 Qh6 7.d3 Nh5 8.Nh4 Qg5 9.Nf5 c6 10.g4 Nf6";

    List<string> _moves = new();
    private int _index = 0;
    
    private ChessBook()
    {
        _chessBoard = new ChessBoard();
        ParsingGame();
    }

    private void ParsingGame()
    {
        var tokens = chessGame.Split(' ');
        foreach (var token in tokens)
        {
            var move = token[(token.IndexOf('.') + 1)..];
            _moves.Add(move);
        }
    }

    private void Print()
    {
        _chessBoard.Print();
    }
    
    public static void Main()
    {
        Console.OutputEncoding = System.Text.Encoding.UTF8;
        var chessBook = new ChessBook();
        chessBook.Print();
        Console.WriteLine("\nN: Next / B: Back / Q,E: Exit");

        var isEndGame = false;
        while (!isEndGame && !chessBook.IsEnd())
        {
            var key = Console.ReadKey(true);
            switch (key.Key)
            {
                case ConsoleKey.Q:
                case ConsoleKey.E:
                     /* pass through */
                case ConsoleKey.Escape:
                    isEndGame = true;
                    break;
                case ConsoleKey.B:
                    chessBook.MoveBack();
                    chessBook.Print();
                    Console.WriteLine("\nN: Next / B: Back / Q,E: Exit");
                    break;
                case ConsoleKey.N:
                    chessBook.MoveNext();
                    chessBook.Print();
                    Console.WriteLine("\nN: Next / B: Back / Q,E: Exit");
                    break;
            }
        }
        Console.WriteLine($"Exit!");
    }

    private bool IsEnd()
    {
        return _index == _moves.Count;
    }

    private void MoveNext()
    {
        if (_index == _moves.Count) return;
        
        var move = _moves[_index++];
        var piece = "";

        if (!"KQBNR".Contains(move[0]))
        {
            piece = "P";
        }
        _chessBoard.MoveNext(piece + move);
    }

    private void MoveBack()
    {
        _chessBoard.MoveBack();
    }
}
