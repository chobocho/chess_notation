class ChessBook
{
    private ChessBoard chessBoard;

    private string chessGame =
        "1.e4 e5 2.f4 exf4 3.Bc4 Qh4+ 4.Kf1 b5 5.Bxb5 Nf6 6.Nf3 Qh6 7.d3 Nh5 8.Nh4 Qg5 9.Nf5 c6 10.g4 Nf6";

    List<string> moves = new();
    private int index = 0;
    
    ChessBook()
    {
        chessBoard = new ChessBoard();
        ParsingGame();
    }

    private void ParsingGame()
    {
        string[] tokens = chessGame.Split(' ');
        foreach (string token in tokens)
        {
            string move = token.Substring(token.IndexOf('.') + 1);
            moves.Add(move);
        }
    }

    private void print()
    {
        chessBoard.Print();
    }
    
    public static void Main()
    {
        Console.OutputEncoding = System.Text.Encoding.UTF8;
        ChessBook chessBook = new ChessBook();
        chessBook.print();
        Console.WriteLine("\nN: Next / B: Back / Q,E: Exit");

        bool isEndGame = false;
        while (!isEndGame && !chessBook.isEnd())
        {
            ConsoleKeyInfo key = Console.ReadKey(true);
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
                    chessBook.print();
                    Console.WriteLine("\nN: Next / B: Back / Q,E: Exit");
                    break;
                case ConsoleKey.N:
                    chessBook.MoveNext();
                    chessBook.print();
                    Console.WriteLine("\nN: Next / B: Back / Q,E: Exit");
                    break;
                default:
                    break;
            }
        }
        Console.WriteLine($"Exit!");
    }

    private bool isEnd()
    {
        return index == moves.Count;
    }

    private void MoveNext()
    {
        if (index == moves.Count) return;
        
        string move = moves[index++];
        string piece = "";

        if (!"KQBNR".Contains(move[0]))
        {
            piece = "P";
        }
        chessBoard.MoveNext(piece + move);
    }

    private void MoveBack()
    {
        chessBoard.MoveBack();
    }
}
