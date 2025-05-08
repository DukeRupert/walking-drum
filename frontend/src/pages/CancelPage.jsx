const CancelPage = () => (
  <div className="cancel-page">
    <h2>Subscription Canceled</h2>
    <p>Your subscription checkout was canceled.</p>
    <button onClick={() => window.location.href = '/'}>Return Home</button>
  </div>
);

export default CancelPage;