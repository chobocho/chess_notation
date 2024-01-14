class ChessBook
{
    private ChessBoard chessBoard;

    ChessBook()
    {
        chessBoard = new ChessBoard();
    }

    private void print()
    {
        chessBoard.print();
    }
    
    public static void Main()
    {
        Console.WriteLine("ChessBook V0.1");
        // Read PGN file
        // N : Next
        // B: Back
        // Q : Exit
        // H : Help
        // S : Start again
        ChessBook chessBook = new ChessBook();
        chessBook.print();
    }

}
